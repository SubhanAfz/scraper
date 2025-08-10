package conversion

type ConversionService interface {
	Convert(htmlContent string) (string, error) // convert html to whatever format, depending on service.
}
