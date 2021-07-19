package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// TODO - proxy ExecuteTemplate method to prevent writing directly to http response writer
// (negates superfluous WriteHeader call when invoking internalError(...).writeTo(w))
type Templates struct{ *template.Template }

var templateFunctions = template.FuncMap{
	"timestamp_as_unix": func(ts time.Time) int64 {
		var emptyTime time.Time

		if ts.Equal(emptyTime) {
			return 0
		}
		return ts.Unix()
	},
	"format_timestamp": func(ts time.Time, layout string) string {
		return ts.Format(layout)
	},
	"jsonify_strings": func(input []string) string {
		bytes, err := json.Marshal(input)
		if err != nil {
			return ""
		}

		return string(bytes)
	},
}

// ParseTemplates parses our HTML templates and returns them collectively for use
func ParseTemplates(viewsPath string) (*Templates, error) {
	// prepare the templates
	tpl := template.New("prediction-league").Funcs(templateFunctions)

	if err := walkPathAndParseTemplates(tpl, fmt.Sprintf("%s/page", viewsPath)); err != nil {
		return nil, fmt.Errorf("cannot walk page path: %w", err)
	}
	if err := walkPathAndParseTemplates(tpl, fmt.Sprintf("%s/email", viewsPath)); err != nil {
		return nil, fmt.Errorf("cannot walk email path: %w", err)
	}

	// return our wrapped template struct
	return &Templates{Template: tpl}, nil
}

// walkPathAndParseTemplates recursively parses templates within a given top-level directory
func walkPathAndParseTemplates(tpl *template.Template, path string) error {
	// walk through our views folder and parse each item to pack the assets
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// we already have an error from a recursive call, so just return with that
		if err != nil {
			return err
		}

		// skip directories, we're only interested in files
		if info.IsDir() {
			return nil
		}

		// open the current file
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		file := bytes.NewReader(contents)

		// copy file contents as a byte stream and then parse as a template
		var b bytes.Buffer
		if _, err = io.Copy(&b, file); err != nil {
			return err
		}
		tpl = template.Must(tpl.Parse(b.String()))

		return nil
	})
}
