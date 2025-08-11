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

	extensions_path := "extensions/autoconsent"

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.UserAgent("'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36'"),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-extensions-except", extensions_path),
		chromedp.Flag("load-extension", extensions_path),
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

func (c *Chrome) ScreenShot() ([]byte, error) {
	var buf []byte

	err := chromedp.Run(c.ctx,
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (c *Chrome) GetPage(url string, waitTime time.Duration) (string, error) {
	var content string

	err := chromedp.Run(c.ctx,
		bypass_webdriver_detection(),
		chromedp.Navigate(url),
		chromedp.Sleep(waitTime),
		get_visible_html(&content),
	)

	if err != nil {
		return "", err
	}

	return content, nil
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
