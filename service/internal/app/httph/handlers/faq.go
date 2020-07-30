package handlers

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"prediction-league/service/internal/pages"
)

func getFAQPageData(realmName string) pages.FAQPageData {
	var data pages.FAQPageData

	// read faq for provided realm
	contents, err := ioutil.ReadFile(fmt.Sprintf("./data/faq-by-realm/%s.yml", realmName))
	if err != nil {
		data.Err = err
		return data
	}

	// parse faqs
	var faqs []pages.FAQItem
	if err := yaml.Unmarshal(contents, &faqs); err != nil {
		data.Err = err
		return data
	}

	// markdownify the answers
	for idx := range faqs {
		bytes := []byte(faqs[idx].Answer)
		html := markdown.ToHTML(bytes, nil, nil)
		faqs[idx].Answer = template.HTML(string(html))
	}

	data.FAQs = faqs

	return data
}
