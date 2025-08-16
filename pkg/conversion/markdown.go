package conversion

import (
	"bytes"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/utils"
	"golang.org/x/net/html"
)

type MarkdownService struct {
	md *htmltomarkdown.Converter
}

func init() {
	Register("markdown", NewMarkdownService())
}

func NewMarkdownService() *MarkdownService {
	return &MarkdownService{
		md: htmltomarkdown.NewConverter(htmltomarkdown.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(
				commonmark.WithStrongDelimiter("__"),
				// ...additional configurations for the plugin
			),
			// ...additional plugins (e.g. table)
		)),
	}
}

func (mdservice *MarkdownService) Convert(page browser.Page) (browser.Page, error) {
	prependedContent, err := prependHrefBaseURL(page.Content, page.URL)
	if err != nil {
		return browser.Page{}, err
	}
	mdContent, err := mdservice.md.ConvertString(prependedContent)
	if err != nil {
		return browser.Page{}, err
	}

	mdContent = utils.RemoveBase64Images(mdContent)

	page.Content = mdContent
	return page, nil
}

func prependHrefBaseURL(htmlString, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, a := range n.Attr {
				if (a.Key == "href" || a.Key == "src") && strings.HasPrefix(a.Val, "/") {
					cleanBaseURL := strings.TrimRight(baseURL, "/")
					n.Attr[i].Val = cleanBaseURL + a.Val
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// Re-render the modified HTML tree to a string.
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
