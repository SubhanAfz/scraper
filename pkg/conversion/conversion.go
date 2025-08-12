package conversion

import (
	"github.com/SubhanAfz/scraper/pkg/browser"
)

type ConversionService interface {
	Convert(page browser.Page) (string, error) // convert content to whatever format, depending on service.
}
