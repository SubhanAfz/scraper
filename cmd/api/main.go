package main

import (
	"net/http"

	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/server"
)

func main() {
	ChromeService, err := browser.NewChrome()
	if err != nil {
		panic(err)
	}
	defer ChromeService.Close()

	server := &server.Server{
		BrowserService: ChromeService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /get_page", server.GetPageHandler)
	mux.HandleFunc("GET /screenshot", server.ScreenShotHandler)
	http.ListenAndServe(":8080", mux)
}
