package conversion

import (
	"github.com/SubhanAfz/scraper/pkg/browser"
)

type ConversionService interface {
	Convert(page browser.Page) (browser.Page, error) // convert content to whatever format, depending on service.
}

var registry = map[string]ConversionService{}

func Register(name string, service ConversionService) {
	registry[name] = service
}

func GetService(name string) (ConversionService, bool) {
	service, exists := registry[name]
	return service, exists
}
