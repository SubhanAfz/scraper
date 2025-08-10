package conversion

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
)

type MarkdownService struct {
	md *htmltomarkdown.Converter
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

func (mdservice *MarkdownService) Convert(htmlContent string) (string, error) {
	return mdservice.md.ConvertString(htmlContent)
}
