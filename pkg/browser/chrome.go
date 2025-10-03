package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/SubhanAfz/scraper/pkg/autoconsent"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Chrome struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewChrome() (*Chrome, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
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
	var url string

	err := chromedp.Run(c.ctx,
		bypass_webdriver_detection(),
		chromedp.Navigate(req.URL),
		chromedp.Sleep(wait),
		chromedp.Location(&url),
	)

	if err != nil {
		return GetScreenShotResponse{}, err
	}
	rule := get_right_rule(c.ctx, url)
	fmt.Printf("Detected rule: %+v\n", rule)
	opt_out(c.ctx, rule)
	fmt.Println("Opt-out actions executed")
	err = chromedp.Run(c.ctx,
		chromedp.FullScreenshot(&buf, 90),
	)

	return GetScreenShotResponse{
		Image: buf,
	}, nil
}

func (c *Chrome) GetPage(req GetPage) (Page, error) {
	var content string
	var title string

	wait := time.Duration(req.WaitTime) * time.Millisecond
	var url string

	err := chromedp.Run(c.ctx,
		bypass_webdriver_detection(),
		chromedp.Navigate(req.URL),
		chromedp.Sleep(wait),
		chromedp.Location(&url),
	)
	if err != nil {
		return Page{}, err
	}
	rule := get_right_rule(c.ctx, url)
	fmt.Printf("Detected rule: %+v\n", rule)
	opt_out(c.ctx, rule)
	fmt.Println("Opt-out actions executed")
	err = chromedp.Run(c.ctx,
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

func get_right_rule(ctx context.Context, url string) autoconsent.AutoConsentRule {
	for _, rule := range autoconsent.Rules.Rules {
		var rightRule bool = true
		if len(rule.DetectCMP) == 0 {
			rightRule = false
		}
		if rule.RunContext.UrlPattern != "" && !rule.RunContext.URLMatches(url) {
			rightRule = false
		}
		if rule.RunContext.UrlPattern != "" && rightRule {
			fmt.Println(rule.RunContext.UrlPattern)
		}
		executed_right := ExecuteActions(ctx, rule.DetectCMP, ModeDetect)
		if !executed_right {
			rightRule = false
		}
		if rightRule {
			return rule
		}
	}
	return autoconsent.AutoConsentRule{}
}

func opt_out(ctx context.Context, rule autoconsent.AutoConsentRule) {
	ExecuteActions(ctx, rule.OptOut, ModeExecute)
}

type ActionMode int

const (
	ModeDetect ActionMode = iota
	ModeExecute
)

func ExecuteActions(ctx context.Context, actions autoconsent.ActionList, mode ActionMode) bool {
	var executed_right = true
	for _, action := range actions {
		fmt.Println("Executing action:", action.ActionType())
		switch a := action.(type) {
		case autoconsent.ClickAction:
			if mode == ModeExecute {
				if err := a.Click.Click(ctx); err != nil {
					fmt.Printf("Error executing click action: %v\n", err)
				}
			}
		case autoconsent.WaitForThenClickAction:
			if mode == ModeExecute {
				if err := a.WaitForClick(ctx); err != nil {
					fmt.Printf("Error waiting for then click action: %v\n", err)
				}
			}
		case autoconsent.ExistsAction:
			if mode == ModeDetect {
				if exists, err := a.Exists.ElementExists(ctx); err != nil || !exists {
					fmt.Printf("Error checking existence for exists action: %v\n", err)
					executed_right = false
					continue
				}
			}
		case autoconsent.VisibleAction:
			if mode == ModeDetect {
				if visible, err := a.Visible.ElementExists(ctx); err != nil || !visible {
					fmt.Printf("Error checking visibility for visible action: %v\n", err)
					executed_right = false
					continue
				}
			}
		case autoconsent.WaitForAction:
			if mode == ModeDetect {
				if exists := a.Wait(ctx); !exists {
					executed_right = false
					continue
				}
			}
		case autoconsent.WaitForVisibleAction:
			if mode == ModeDetect {
				if visible := a.Wait(ctx); !visible {
					executed_right = false
					continue
				}
			}
		case autoconsent.IfThenElseAction:
			if mode == ModeExecute {
				// Create a new ActionList containing the single 'if' action
				fmt.Println(a.If.ActionType())
				ifActions := autoconsent.ActionList{a.If}
				if ExecuteActions(ctx, ifActions, ModeDetect) {
					fmt.Println("Condition met for IfThenElseAction")
					// If condition is true, execute 'then' actions
					ExecuteActions(ctx, a.Then, ModeExecute)
				} else {
					// If condition is false, execute 'else' actions (if any)
					fmt.Println("Condition not met for IfThenElseAction")
					ExecuteActions(ctx, a.Else, ModeExecute)
				}
			}
		case autoconsent.UnconditionalWaitAction:
			a.Wait(ctx)
		case autoconsent.EvalAction:
			if mode == ModeExecute {
				a.Evaluate(ctx)
			}
		default:
			fmt.Printf("Unsupported action type detected: %T\n", a)
		}
	}
	return executed_right
}
