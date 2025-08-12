package browser

import "time"

type Page struct {
	Title   string
	Content string
	URL     string
}

// BrowserService defines the interface for browser operations
type BrowserService interface {
	Close()                                                   // closes the browser instance
	GetPage(url string, waitTime time.Duration) (Page, error) // gets the HTML content of a page
	ScreenShot() ([]byte, error)                              // takes a full screenshot of the current page
}
