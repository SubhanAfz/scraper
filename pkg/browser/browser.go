package browser

import "time"

// BrowserService defines the interface for browser operations
type BrowserService interface {
	Close()                                                     // closes the browser instance
	GetPage(url string, waitTime time.Duration) (string, error) // gets the HTML content of a page
	ScreenShot() ([]byte, error)                                // takes a full screenshot of the current page
}
