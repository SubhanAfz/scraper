package autoconsent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type AutoConsentRules struct {
	Rules []AutoConsentRule `json:"autoconsent"`
}

type AutoConsentRule struct {
	Name        string     `json:"name"`
	DetectCMP   ActionList `json:"detectCMP"`
	DetectPopup ActionList `json:"detectPopup"`
	OptIn       ActionList `json:"optIn"`
	OptOut      ActionList `json:"optOut"`
	RunContext  RunContext `json:"runContext"`
}

type RunContext struct {
	UrlPattern string `json:"urlPattern"`
}

func (r *RunContext) URLMatches(url string) bool {
	if r.UrlPattern == "" {
		return true // No pattern means it matches all URLs
	}
	re := regexp.MustCompile(r.UrlPattern)
	return re.MatchString(url)
}

type ElementSelector struct {
	Element interface{}
}

func (e *ElementSelector) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Element = s
		return nil
	}
	// Try to unmarshal as []string
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		e.Element = arr
		return nil
	}
	return fmt.Errorf("ElementSelector must be string or []string")
}

func (e *ElementSelector) ElementExists(ctx context.Context) (bool, error) {
	switch s := e.Element.(type) {
	case string:
		if strings.HasPrefix(s, "xpath/") {
			return elementExistsXPath(ctx, s[6:])
		}
		return elementExists(ctx, s)
	case []string:
		return elementExistsComplex(ctx, s)
	default:
		return false, fmt.Errorf("unsupported selector type: %T", e.Element)
	}
}

func (e *ElementSelector) Click(ctx context.Context) error {
	switch s := e.Element.(type) {
	case string:
		if strings.HasPrefix(s, "xpath/") {
			return chromedp.Run(ctx, chromedp.Click(s[6:], chromedp.BySearch))
		}
		return chromedp.Run(ctx, chromedp.Click(s))
	case []string:
		return clickComplex(ctx, s)
	default:
		return fmt.Errorf("unsupported selector type for click: %T", e.Element)
	}
}

func clickComplex(ctx context.Context, selectors []string) error {
	// Build JavaScript to handle shadow DOM and iframe piercing for clicking
	selectorsJSON, err := json.Marshal(selectors)
	if err != nil {
		return err
	}

	js := fmt.Sprintf(`
        (function() {
            function clickElement(selectors) {
                let element = document;
                
                for (let i = 0; i < selectors.length; i++) {
                    const selector = selectors[i];
                    
                    if (selector.startsWith('xpath/')) {
                        // Handle XPath - evaluate against current context
                        const xpath = selector.substring(6);
                        const contextNode = element.nodeType === Node.DOCUMENT_NODE || element.nodeType === Node.DOCUMENT_FRAGMENT_NODE ? element : element.ownerDocument || document;
                        const result = contextNode.evaluate(xpath, element, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);
                        element = result.singleNodeValue;
                    } else {
                        // Handle CSS selector
                        element = element.querySelector(selector);
                    }
                    
                    if (!element) return false;
                    
                    // If this is the last selector, click the element
                    if (i === selectors.length - 1) {
                        element.click();
                        return true;
                    }
                    
                    // Pierce shadow DOM if available
                    if (element.shadowRoot) {
                        element = element.shadowRoot;
                    }
                    // Pierce iframe if available and same-origin
                    else if (element.tagName === 'IFRAME' && element.contentDocument) {
                        element = element.contentDocument;
                    }
                }
                
                return false;
            }
            
            return clickElement(%s);
        })()`, string(selectorsJSON))

	var success bool
	err = chromedp.Run(ctx, chromedp.Evaluate(js, &success))
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("failed to click element with complex selector")
	}
	return nil
}

func elementExists(ctx context.Context, selector string) (bool, error) {
	var exists bool
	js := fmt.Sprintf(`document.querySelector(%s) !== null`, jsonEscape(selector))
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &exists))
	return exists, err
}

func elementExistsXPath(ctx context.Context, xpath string) (bool, error) {
	var exists bool
	js := fmt.Sprintf(`document.evaluate(%s, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue !== null`, jsonEscape(xpath))
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &exists))
	return exists, err
}

func jsonEscape(s string) string {
	escaped, _ := json.Marshal(s)
	return string(escaped)
}

func elementExistsComplex(ctx context.Context, selectors []string) (bool, error) {
	// Start with the first selector
	var exists bool

	// Build JavaScript to handle shadow DOM and iframe piercing
	selectorsJSON, err := json.Marshal(selectors)
	if err != nil {
		return false, err
	}

	js := fmt.Sprintf(`
        (function() {
            function findElement(selectors) {
                let element = document;
                
                for (let i = 0; i < selectors.length; i++) {
                    const selector = selectors[i];
                    
                    if (selector.startsWith('xpath/')) {
                        // Handle XPath - evaluate against current context (document, shadowRoot, or iframe document)
                        const xpath = selector.substring(6);
                        const contextNode = element.nodeType === Node.DOCUMENT_NODE || element.nodeType === Node.DOCUMENT_FRAGMENT_NODE ? element : element.ownerDocument || document;
                        const result = contextNode.evaluate(xpath, element, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);
                        element = result.singleNodeValue;
                    } else {
                        // Handle CSS selector - works on Document, Element, and ShadowRoot
                        element = element.querySelector(selector);
                    }
                    
                    if (!element) return false;
                    
                    // Pierce shadow DOM if available
                    if (element.shadowRoot) {
                        element = element.shadowRoot;
                    }
                    // Pierce iframe if available and same-origin
                    else if (element.tagName === 'IFRAME' && element.contentDocument) {
                        element = element.contentDocument;
                    }
                }
                
                return element !== null && element !== document;
            }
            
            return findElement(%s);
        })()`, string(selectorsJSON))

	err = chromedp.Run(ctx, chromedp.Evaluate(js, &exists))
	return exists, err
}

type Action interface {
	ActionType() string
}

type ExistsAction struct {
	Exists ElementSelector `json:"exists"`
}

func (e ExistsAction) ActionType() string {
	return "exists"
}

type VisibleAction struct {
	Visible ElementSelector `json:"visible"`
	Check   string          `json:"check"`
}

func (v VisibleAction) ActionType() string {
	return "visible"
}

type WaitForAction struct {
	WaitFor ElementSelector `json:"waitFor"`
	Timeout uint64          `json:"timeout"`
}

func (w WaitForAction) Wait(ctx context.Context) bool {
	interval := 100 * time.Millisecond
	deadline := time.Now().Add(time.Duration(w.Timeout) * time.Millisecond)
	for time.Now().Before(deadline) {
		exists, err := w.WaitFor.ElementExists(ctx)
		if err == nil && exists {
			return true
		}
		chromedp.Run(ctx, chromedp.Sleep(interval))
	}

	return false
}

func (w WaitForAction) ActionType() string {
	return "waitFor"
}

type WaitForVisibleAction struct {
	WaitFor ElementSelector `json:"waitForVisible"`
	Timeout uint64          `json:"timeout"`
	Check   string          `json:"check"`
}

func (w WaitForVisibleAction) Wait(ctx context.Context) bool {
	interval := 100 * time.Millisecond
	deadline := time.Now().Add(time.Duration(w.Timeout) * time.Millisecond)
	for time.Now().Before(deadline) {
		exists, err := w.WaitFor.ElementExists(ctx)
		if err == nil && exists {
			return true
		}
		chromedp.Run(ctx, chromedp.Sleep(interval))
	}

	return false
}

func (w WaitForVisibleAction) ActionType() string {
	return "waitForVisible"
}

type ClickAction struct {
	Click ElementSelector `json:"click"`
	All   bool            `json:"all"`
}

func (c ClickAction) ActionType() string {
	return "click"
}

type WaitForThenClickAction struct {
	WaitFor ElementSelector `json:"waitForThenClick"`
	Timeout uint64          `json:"timeout"`
	Check   string          `json:"check"`
}

func (w WaitForThenClickAction) ActionType() string {
	return "waitForThenClick"
}

func (w WaitForThenClickAction) WaitForClick(ctx context.Context) error {
	interval := 100 * time.Millisecond
	deadline := time.Now().Add(time.Duration(w.Timeout) * time.Millisecond)
	for time.Now().Before(deadline) {
		exists, err := w.WaitFor.ElementExists(ctx)
		if err == nil && exists {
			// Element found, click it and return immediately
			return w.WaitFor.Click(ctx)
		}
		chromedp.Run(ctx, chromedp.Sleep(interval))
	}
	// Timeout reached without finding element
	return fmt.Errorf("timeout waiting for element to appear")
}

type HideAction struct {
	Hide   string `json:"hide"`
	Method string `json:"method"`
}

func (h HideAction) ActionType() string {
	return "hide"
}

type CookieMatchAction struct {
	CookieContains string `json:"cookieContains"`
}

func (c CookieMatchAction) ActionType() string {
	return "cookieMatch"
}

type EvalAction struct {
	Eval string `json:"eval"`
}

func (e EvalAction) ActionType() string {
	return "eval"
}

func (e EvalAction) Evaluate(ctx context.Context) error {
	return chromedp.Run(ctx, chromedp.Evaluate("(()=>{"+JSEvals[e.Eval]+"})()", nil))
}

type IfThenElseAction struct {
	If   Action     `json:"if"`
	Then ActionList `json:"then"`
	Else ActionList `json:"else"`
}

func (i IfThenElseAction) ActionType() string {
	return "ifThenElse"
}

type UnconditionalWaitAction struct {
	WaitTime uint64 `json:"wait"`
}

func (u UnconditionalWaitAction) ActionType() string {
	return "wait"
}

func (u UnconditionalWaitAction) Wait(ctx context.Context) {
	chromedp.Run(ctx, chromedp.Sleep(time.Duration(u.WaitTime)*time.Millisecond))
}

type ActionList []Action

func (al *ActionList) UnmarshalJSON(data []byte) error {
	var rawList []map[string]interface{}
	if err := json.Unmarshal(data, &rawList); err != nil {
		return err
	}
	var result []Action
	for _, raw := range rawList {
		var act Action
		switch {
		case raw["exists"] != nil:
			var a ExistsAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["visible"] != nil:
			var a VisibleAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["waitFor"] != nil:
			var a WaitForAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			// Set default timeout if not specified
			if a.Timeout == 0 {
				a.Timeout = 1000
			}
			act = a
		case raw["waitForVisible"] != nil:
			var a WaitForVisibleAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			// Set default timeout if not specified
			if a.Timeout == 0 {
				a.Timeout = 1000
			}
			act = a
		case raw["click"] != nil:
			var a ClickAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["waitForThenClick"] != nil:
			var a WaitForThenClickAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			// Set default timeout if not specified
			if a.Timeout == 0 {
				a.Timeout = 1000
			}
			act = a
		case raw["hide"] != nil:
			var a HideAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["cookieContains"] != nil:
			var a CookieMatchAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["eval"] != nil:
			var a EvalAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["wait"] != nil:
			var a UnconditionalWaitAction
			b, _ := json.Marshal(raw)
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			act = a
		case raw["if"] != nil:
			var a IfThenElseAction

			// Handle the nested structure manually
			var rawIf map[string]interface{}
			if ifVal, ok := raw["if"].(map[string]interface{}); ok {
				rawIf = ifVal
			} else {
				return fmt.Errorf("if field must be an object")
			}

			// Parse the "if" action
			var ifAction Action
			switch {
			case rawIf["exists"] != nil:
				var ea ExistsAction
				ifBytes, _ := json.Marshal(rawIf)
				if err := json.Unmarshal(ifBytes, &ea); err != nil {
					return err
				}
				ifAction = ea
			case rawIf["visible"] != nil:
				var va VisibleAction
				ifBytes, _ := json.Marshal(rawIf)
				if err := json.Unmarshal(ifBytes, &va); err != nil {
					return err
				}
				ifAction = va
			default:
				return fmt.Errorf("unsupported if action type: %+v", rawIf)
			}

			a.If = ifAction

			// Parse "then" and "else" normally
			if then, ok := raw["then"]; ok {
				thenBytes, _ := json.Marshal(then)
				if err := json.Unmarshal(thenBytes, &a.Then); err != nil {
					return err
				}
			}
			if else_, ok := raw["else"]; ok {
				elseBytes, _ := json.Marshal(else_)
				if err := json.Unmarshal(elseBytes, &a.Else); err != nil {
					return err
				}
			}

			act = a
		default:
			continue
			//return fmt.Errorf("unknown action type: %+v", raw)
		}
		result = append(result, act)
	}
	*al = result
	return nil
}
