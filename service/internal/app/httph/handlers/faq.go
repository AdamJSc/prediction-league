package handlers

import (
	"github.com/gomarkdown/markdown"
	"html/template"
	"prediction-league/service/internal/pages"
)

func getFAQPageData(realmName string) pages.FAQPageData {
	var data pages.FAQPageData

	faqs := []pages.FAQItem{
		{
			Question: "Which would you rather bee or a wasp?",
			Answer: `That depends on what we mean by **bee**.

In the traditional sense, a bee would be far more aware of its mortality following a single sting,
so would logically cause _fewer_ injuries to [humans](https://en.wikipedia.org/wiki/Human).

So, on the grounds of compassion, a bee.
`,
		},
		{
			Question: "What is the average air speed velocity of a laden swallow?",
			Answer: `The real question is "_how did the coconut get to medieval England?_"`,
		},
		{
			Question: "What's the Frequency, Kenneth?",
			Answer: `"_Somewhere between 88.1 and 90.2 MHz on FM_" - **Mr Bruce**`,
		},
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
