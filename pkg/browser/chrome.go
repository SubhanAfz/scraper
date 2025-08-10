package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Chrome struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewChrome() (*Chrome, error) {
	ctx, cancel := chromedp.NewContext(context.Background())

	err := chromedp.Run(ctx, chromedp.Navigate("about:blank"))

	if err != nil {
		cancel()
		return nil, err
	}

	return &Chrome{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (c *Chrome) Close() {
	c.cancel()
}

func (c *Chrome) GetPage(url string, waitTime time.Duration) (string, error) {
	var content string

	err := chromedp.Run(c.ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(waitTime),
		chromedp.OuterHTML("html", &content),
	)

	if err != nil {
		return "", err
	}

	return content, nil
}
