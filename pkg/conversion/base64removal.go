package conversion

import (
	"regexp"

	"github.com/SubhanAfz/scraper/pkg/browser"
)

type Base64RemovalService struct {
	re *regexp.Regexp
}

func NewBase64RemovalService() *Base64RemovalService {
	return &Base64RemovalService{
		re: regexp.MustCompile(`data:image\/([a-zA-Z]*);base64,([a-zA-Z0-9+\/]*={0,2})`), // matches base64 images.
	}
}

func (b *Base64RemovalService) Convert(page browser.Page) (browser.Page, error) {
	page.Content = b.re.ReplaceAllString(page.Content, "Base64 Image Removed")
	return page, nil
}
