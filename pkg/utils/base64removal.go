package utils

import (
	"regexp"
)

func RemoveBase64Images(content string) string {
	re := regexp.MustCompile(`data:image\/([a-zA-Z]*);base64,([a-zA-Z0-9+\/]*={0,2})`)
	return re.ReplaceAllString(content, "Base64 Image Removed")
}
