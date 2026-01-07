package browser

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/SubhanAfz/scraper/pkg/autoconsent"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Chrome struct {
	ctx            context.Context
	ctxCancel      context.CancelFunc
	allocCancel    context.CancelFunc
	tempProfileDir string
}

func NewChrome() (*Chrome, error) {
	return NewChromeWithConfig(DefaultChromeConfig())
}

type ChromeConfig struct {
	Headless            bool
	DisableSandbox      bool
	UserDataDir         string
	ExtensionPaths      []string
	RemoteDebuggingHost string
	RemoteDebuggingPort int
	RemoteAllocatorURL  string
}

func DefaultChromeConfig() ChromeConfig {
	return ChromeConfig{
		Headless: true,
	}
}

func NewChromeWithConfig(cfg ChromeConfig) (*Chrome, error) {
	tempProfileDir := ""
	if cfg.UserDataDir == "" && cfg.RemoteAllocatorURL == "" {
		dir, err := os.MkdirTemp("", "scraper-profile-")
		if err != nil {
			return nil, err
		}
		tempProfileDir = dir
		cfg.UserDataDir = dir
	}

	allocatorCtx, allocCancel := createAllocator(cfg)
	ctx, ctxCancel := chromedp.NewContext(allocatorCtx)

	startupTasks := []chromedp.Action{}
	// Always run stealth scripts at startup
	startupTasks = append(startupTasks, bypass_webdriver_detection())
	startupTasks = append(startupTasks, disableDebuggerDetection())
	if cfg.Headless {
		startupTasks = append(startupTasks, removeHeadlessUserAgent())
	}
	startupTasks = append(startupTasks, chromedp.Navigate("about:blank"))

	err := chromedp.Run(ctx, startupTasks...)
	if err != nil {
		ctxCancel()
		allocCancel()
		if tempProfileDir != "" {
			_ = os.RemoveAll(tempProfileDir)
		}
		return nil, err
	}

	return &Chrome{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		allocCancel:    allocCancel,
		tempProfileDir: tempProfileDir,
	}, nil
}

func createAllocator(cfg ChromeConfig) (context.Context, context.CancelFunc) {
	if cfg.RemoteAllocatorURL != "" {
		return chromedp.NewRemoteAllocator(context.Background(), cfg.RemoteAllocatorURL)
	}
	opts := execAllocatorOptions(cfg)
	return chromedp.NewExecAllocator(context.Background(), opts...)
}

func execAllocatorOptions(cfg ChromeConfig) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Default anti-bot-ish flags matching nodriver's Config defaults.
		chromedp.Flag("remote-allow-origins", "*"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-service-autorun", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("homepage", "about:blank"),
		chromedp.Flag("no-pings", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-session-crashed-bubble", true),
		chromedp.Flag("disable-search-engine-choice-screen", true),
		// Anti-bot detection flags
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("useAutomationExtension", false),
	)

	if cfg.UserDataDir != "" {
		opts = append(opts, chromedp.Flag("user-data-dir", cfg.UserDataDir))
	}

	features := []string{"IsolateOrigins", "site-per-process"}
	if len(cfg.ExtensionPaths) > 0 {
		features = append(features, "DisableLoadExtensionCommandLineSwitch")
		opts = append(opts,
			chromedp.Flag("disable-extensions-except", strings.Join(cfg.ExtensionPaths, ",")),
			chromedp.Flag("load-extension", strings.Join(cfg.ExtensionPaths, ",")),
		)
	}
	opts = append(opts, chromedp.Flag("disable-features", strings.Join(features, ",")))

	if cfg.Headless {
		opts = append(opts, chromedp.Flag("headless", "new"))
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	if cfg.DisableSandbox {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}

	if cfg.RemoteDebuggingHost != "" {
		opts = append(opts, chromedp.Flag("remote-debugging-host", cfg.RemoteDebuggingHost))
	}
	if cfg.RemoteDebuggingPort != 0 {
		opts = append(opts, chromedp.Flag("remote-debugging-port", cfg.RemoteDebuggingPort))
	}

	return opts
}

func (c *Chrome) Close() {
	c.ctxCancel()
	c.allocCancel()
	if c.tempProfileDir != "" {
		_ = os.RemoveAll(c.tempProfileDir)
	}
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
	opt_out(c.ctx, rule)
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
	opt_out(c.ctx, rule)
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
		// Comprehensive stealth script to bypass bot detection
		stealthScript := `
		// Remove webdriver property completely by redefining Navigator prototype
		const newProto = Navigator.prototype;
		delete newProto.webdriver;
		
		// Also ensure navigator.webdriver returns undefined naturally
		if (Object.getOwnPropertyDescriptor(navigator, 'webdriver')) {
			Object.defineProperty(navigator, 'webdriver', {
				get: () => undefined,
				configurable: true
			});
		}
		
		// Remove CDP artifacts (cdc_ variables)
		const cdcKeys = Object.keys(window).filter(key => /^cdc_|^__cdc_/.test(key));
		cdcKeys.forEach(key => { try { delete window[key]; } catch(e) {} });
		
		// Override console methods to defeat DevTools timing detection
		// DevTools detection measures how long console.log/table take to execute
		// When DevTools is open, these take much longer due to rendering
		// By making them no-ops, the timing check fails to detect DevTools
		const originalConsole = {
			log: console.log,
			table: console.table,
			clear: console.clear
		};
		console.log = function() { return undefined; };
		console.table = function() { return undefined; };
		console.clear = function() { return undefined; };
		
		// Block devtoolsFormatters detection
		// Chrome calls custom formatters when displaying objects in DevTools console
		Object.defineProperty(window, 'devtoolsFormatters', {
			get: () => undefined,
			set: () => {},
			configurable: false
		});
		
		// Intercept Function constructor to neutralize debugger statement detection
		// Some detection scripts create functions with 'debugger' to measure pause time
		const OriginalFunction = window.Function;
		window.Function = function(...args) {
			if (args.length > 0) {
				const lastArg = args[args.length - 1];
				if (typeof lastArg === 'string' && lastArg.includes('debugger')) {
					args[args.length - 1] = lastArg.replace(/debugger/gi, '');
				}
			}
			return OriginalFunction.apply(this, args);
		};
		window.Function.prototype = OriginalFunction.prototype;
		Object.defineProperty(window.Function, 'name', { value: 'Function' });
		`
		_, err := page.AddScriptToEvaluateOnNewDocument(stealthScript).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

// disableDebuggerDetection disables debugger statement pausing via CDP
// This prevents timing-based detection that measures how long 'debugger' statements take
func disableDebuggerDetection() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Enable the debugger domain first
		_, err := debugger.Enable().Do(ctx)
		if err != nil {
			return err
		}
		// Skip all pauses - this makes 'debugger' statements execute instantly
		return debugger.SetSkipAllPauses(true).Do(ctx)
	})
}

func removeHeadlessUserAgent() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := network.Enable().Do(ctx); err != nil {
			return err
		}
		resp, _, err := runtime.Evaluate(`navigator.userAgent`).WithReturnByValue(true).Do(ctx)
		if err != nil {
			return err
		}
		if resp == nil {
			return nil
		}
		var ua string
		if err := json.Unmarshal(resp.Value, &ua); err != nil {
			return err
		}
		cleanedUA := strings.ReplaceAll(ua, "Headless", "")
		if cleanedUA == ua {
			return nil
		}
		return emulation.SetUserAgentOverride(cleanedUA).Do(ctx)
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
		switch a := action.(type) {
		case autoconsent.ClickAction:
			if mode == ModeExecute {
				if err := a.Click.Click(ctx); err != nil {
				}
			}
		case autoconsent.WaitForThenClickAction:
			if mode == ModeExecute {
				if err := a.WaitForClick(ctx); err != nil {
				}
			}
		case autoconsent.ExistsAction:
			if mode == ModeDetect {
				if exists, err := a.Exists.ElementExists(ctx); err != nil || !exists {
					executed_right = false
					continue
				}
			}
		case autoconsent.VisibleAction:
			if mode == ModeDetect {
				if visible, err := a.Visible.ElementExists(ctx); err != nil || !visible {
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
				ifActions := autoconsent.ActionList{a.If}
				if ExecuteActions(ctx, ifActions, ModeDetect) {
					// If condition is true, execute 'then' actions
					ExecuteActions(ctx, a.Then, ModeExecute)
				} else {
					// If condition is false, execute 'else' actions (if any)
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
		}
	}
	return executed_right
}
