package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"prediction-league/service/internal/views"
	"time"
)

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

// MustParseTemplates parses our HTML templates and returns them collectively for use
func MustParseTemplates(viewsPath string) *views.Templates {
	// prepare the templates
	tpl := template.New("prediction-league").Funcs(templateFunctions)

	mustWalkPathAndParseTemplates(tpl, fmt.Sprintf("%s/page", viewsPath))
	mustWalkPathAndParseTemplates(tpl, fmt.Sprintf("%s/email", viewsPath))

	// return our wrapped template struct
	return &views.Templates{Template: tpl}
}

// mustWalkPathAndParseTemplates recursively parses templates within a given top-level directory
func mustWalkPathAndParseTemplates(tpl *template.Template, path string) {
	// walk through our views folder and parse each item to pack the assets
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
	if err != nil {
		log.Fatal(err)
	}
}
