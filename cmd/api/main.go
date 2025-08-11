package main

import (
	"fmt"
	"log"
	"os"
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

	content, err := BrowserService.GetPage("https://www.google.com", 1500*time.Millisecond)
	if err != nil {
		panic(err)
	}

	mdContent, err := MarkdownService.Convert(content)
	if err != nil {
		panic(err)
	}

	fmt.Println(mdContent)

	screenshot, err := BrowserService.ScreenShot()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("fullScreenshot.png", screenshot, 0o644); err != nil {
		log.Fatal(err)
	}

}
