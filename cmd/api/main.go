package main

import (
	"fmt"
	"time"

	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/conversion"
)

func main() {
	BrowserService, err := browser.NewChrome()
	MarkdownService := conversion.NewMarkdownService()
	Base64RemovalService := conversion.NewBase64RemovalService()

	if err != nil {
		panic(err)
	}

	defer BrowserService.Close()

	page, err := BrowserService.GetPage("https://roblox.com", 500*time.Millisecond)
	if err != nil {
		panic(err)
	}

	mdPage, err := MarkdownService.Convert(page)
	if err != nil {
		panic(err)
	}

	base64Page, err := Base64RemovalService.Convert(mdPage)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Title: %s\n\nMarkdown Content:\n %s", page.Title, base64Page.Content)
}
