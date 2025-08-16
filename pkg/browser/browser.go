package browser

/*
Page struct represents a web page.

	title: the title of the page
	content: the visible content of the page
	url: the URL of the page
*/
type Page struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

/*
GetPageRequest represents a request to get a web page.
	title: the URL of the page to retrieve
	wait_time: the time to wait for the page to load in milliseconds
*/

type GetPage struct {
	URL      string `json:"url"`
	WaitTime uint64 `json:"wait_time"`
}

/*
GetScreenShotRequest represents a request to take a screenshot of a web page.
	title: the URL of the page to capture
	wait_time: the time to wait for the page to load in milliseconds
*/

type GetScreenShotRequest struct {
	URL      string `json:"url"`
	WaitTime uint64 `json:"wait_time"`
}

/*
GetScreenShotResponse represents a response to a request for a screenshot of a web page.
	image: the screenshot image data
*/

type GetScreenShotResponse struct {
	Image []byte `json:"image"`
}

// BrowserService defines the interface for browser operations
type BrowserService interface {
	Close()                                                             // closes the browser instance
	GetPage(req GetPage) (Page, error)                                  // gets the HTML content of a page
	ScreenShot(req GetScreenShotRequest) (GetScreenShotResponse, error) // takes a full screenshot of the current page
}
