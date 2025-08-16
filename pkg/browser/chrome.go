package browser

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Chrome struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewChrome() (*Chrome, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.UserAgent("'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36'"),
	)

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, _ := chromedp.NewContext(allocatorCtx)

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

func (c *Chrome) ScreenShot(req GetScreenShotRequest) (GetScreenShotResponse, error) {
	var buf []byte

	wait := time.Duration(req.WaitTime) * time.Millisecond

	err := chromedp.Run(c.ctx,
		bypass_webdriver_detection(),
		chromedp.Navigate(req.URL),
		chromedp.Sleep(wait),
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return GetScreenShotResponse{}, err
	}

	return GetScreenShotResponse{
		Image: buf,
	}, nil
}

func (c *Chrome) GetPage(req GetPage) (Page, error) {
	var content string
	var title string

	wait := time.Duration(req.WaitTime) * time.Millisecond

	err := chromedp.Run(c.ctx,
		bypass_webdriver_detection(),
		chromedp.Navigate(req.URL),
		chromedp.Sleep(wait),
		get_visible_html(&content),
		get_title(&title),
	)

	if err != nil {
		return Page{}, err
	}

	return Page{
		Title:   title,
		Content: content,
		URL:     req.URL,
	}, nil
}

func bypass_webdriver_detection() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(`Object.defineProperty(navigator, 'webdriver', {
    get: () => false,
  });`).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func get_visible_html(visibleHTML *string) chromedp.Action {
	return chromedp.Evaluate(`(() => {
        function isElementVisible(el) {
            if (!el || !document.body.contains(el)) {
                return false;
            }
            let parent = el;
            while (parent && parent !== document.body) {
                const style = window.getComputedStyle(parent);
                if (style.display === 'none') {
                    return false;
                }
                parent = parent.parentElement;
            }
            const style = window.getComputedStyle(el);
            if (style.display === 'none') return false;
            if (style.visibility === 'hidden' || style.visibility === 'collapse') return false;
            if (parseFloat(style.opacity) === 0) return false;
            const rect = el.getBoundingClientRect();
            if ((rect.width === 0 || rect.height === 0) && style.overflow !== 'visible') {
                return false;
            }
            return true;
        }
        function getVisibleHtml() {
            const allElements = document.querySelectorAll('body *');
            const visibleHtmlParts = [];

            allElements.forEach(element => {
                let parent = element.parentElement;
                if (parent && (parent === document.body || !isElementVisible(parent))) {
                    if (isElementVisible(element)) {
                        visibleHtmlParts.push(element.outerHTML);
                    }
                }
            });
            return visibleHtmlParts.join('\n\n');
        }
        return getVisibleHtml();
    })()`, &visibleHTML)
}
func get_title(title *string) chromedp.Action {
	return chromedp.Evaluate(`document.title`, title)
}
