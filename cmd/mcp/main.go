package main

import (
	"log"
	"net/http"
	"os"

	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cfg := browser.DefaultChromeConfig()
	if os.Getenv("HEADLESS") == "true" {
		cfg.DisableSandbox = true
	}

	ChromeService, err := browser.NewChromeWithConfig(cfg)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port
	log.Printf("Starting MCP server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
