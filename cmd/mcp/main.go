package main

import (
	"log"
	"net/http"

	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ChromeService, err := browser.NewChrome()
	if err != nil {
		panic(err)
	}
	defer ChromeService.Close()
	s := &server.Server{
		BrowserService: ChromeService,
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "web_scraper", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "get_page", Description: "Fetch a web page"}, s.GetPageMCPHandler)
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)
	if err := http.ListenAndServe("localhost:8080", handler); err != nil {
		log.Fatal(err)
	}
}
