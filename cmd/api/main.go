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

	if err != nil {
		panic(err)
	}

	defer BrowserService.Close()

	content, err := BrowserService.GetPage("https://www.reddit.com/r/LocalLLaMA/", 1500*time.Millisecond)
	if err != nil {
		panic(err)
	}

	mdContent, err := MarkdownService.Convert(content)
	if err != nil {
		panic(err)
	}

	fmt.Println(mdContent)
}
