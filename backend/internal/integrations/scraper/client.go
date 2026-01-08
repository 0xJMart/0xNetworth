package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/cdp"
)

// Client handles web scraping operations
type Client struct {
	headless bool
	timeout  time.Duration
}

// NewClient creates a new scraper client
func NewClient(headless bool, timeout time.Duration) *Client {
	return &Client{
		headless: headless,
		timeout:  timeout,
	}
}

// SessionCookie represents a browser session cookie
type SessionCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	SameSite string `json:"sameSite"`
}

// AuthenticateResult represents the result of authentication
type AuthenticateResult struct {
	Success      bool
	SessionCookies []SessionCookie
	Error        string
}

// ExtractYouTubeURLsResult represents the result of URL extraction
type ExtractYouTubeURLsResult struct {
	URLs  []string
	Error string
}

// Authenticate handles email + OTP authentication flow
func (c *Client) Authenticate(ctx context.Context, email, otpCode, loginURL string) (*AuthenticateResult, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", c.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var cookies []SessionCookie
	var authSuccess bool
	var authError string

	err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible("input[type='email'], input[name='email'], input[id*='email']", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Fill email
		chromedp.SendKeys("input[type='email'], input[name='email'], input[id*='email']", email+"\n", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Wait for OTP input field
		chromedp.WaitVisible("input[type='text'], input[name*='code'], input[id*='code'], input[type='number']", chromedp.ByQuery),
		// Fill OTP code
		chromedp.SendKeys("input[type='text'], input[name*='code'], input[id*='code'], input[type='number']", otpCode+"\n", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		// Check if we're logged in (look for common post-login indicators)
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Get current URL to check if we've been redirected
			var currentURL string
			if err := chromedp.Location(&currentURL).Do(ctx); err == nil {
				// If we're not on the login page anymore, assume success
				if !strings.Contains(currentURL, "login") && !strings.Contains(currentURL, "auth") {
					authSuccess = true
				}
			}
			return nil
		}),
		// Extract cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookieList, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			for _, cookie := range cookieList {
				var expires int64
				if cookie.Expires > 0 {
					expires = int64(cookie.Expires)
				}
				cookies = append(cookies, SessionCookie{
					Name:     cookie.Name,
					Value:    cookie.Value,
					Domain:   cookie.Domain,
					Path:     cookie.Path,
					Expires:  expires,
					HTTPOnly: cookie.HTTPOnly,
					Secure:   cookie.Secure,
					SameSite: string(cookie.SameSite),
				})
			}
			return nil
		}),
	)

	if err != nil {
		authError = err.Error()
		return &AuthenticateResult{
			Success: false,
			Error:   authError,
		}, err
	}

	if !authSuccess {
		authError = "Authentication failed - could not verify login success"
	}

	return &AuthenticateResult{
		Success:        authSuccess,
		SessionCookies: cookies,
		Error:          authError,
	}, nil
}

// LoadSession restores a session from cookies
func (c *Client) LoadSession(ctx context.Context, cookies []SessionCookie, baseURL string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", c.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Convert SessionCookie to network cookies and set them
	for _, cookie := range cookies {
		var expires *cdp.TimeSinceEpoch
		if cookie.Expires > 0 {
			exp := cdp.TimeSinceEpoch(time.Unix(cookie.Expires, 0))
			expires = &exp
		}
		
		setCookie := network.SetCookie(cookie.Name, cookie.Value).
			WithURL(baseURL).
			WithHTTPOnly(cookie.HTTPOnly).
			WithSecure(cookie.Secure)
		
		if cookie.Domain != "" {
			setCookie = setCookie.WithDomain(cookie.Domain)
		}
		if cookie.Path != "" {
			setCookie = setCookie.WithPath(cookie.Path)
		}
		if expires != nil {
			setCookie = setCookie.WithExpires(expires)
		}
		if cookie.SameSite != "" {
			setCookie = setCookie.WithSameSite(network.CookieSameSite(cookie.SameSite))
		}
		
		if err := setCookie.Do(ctx); err != nil {
			log.Printf("Warning: Failed to set cookie %s: %v", cookie.Name, err)
		}
	}

	return chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.Sleep(2*time.Second),
	)
}

// ExtractYouTubeURLs extracts YouTube video URLs from a page
func (c *Client) ExtractYouTubeURLs(ctx context.Context, pageURL string) (*ExtractYouTubeURLsResult, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", c.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var pageHTML string
	var urls []string

	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for dynamic content to load
		chromedp.OuterHTML("html", &pageHTML),
	)

	if err != nil {
		return &ExtractYouTubeURLsResult{
			URLs:  nil,
			Error: err.Error(),
		}, err
	}

	// Extract YouTube URLs using multiple methods
	urls = c.extractURLsFromHTML(pageHTML)

	// Remove duplicates
	urls = c.removeDuplicates(urls)

	return &ExtractYouTubeURLsResult{
		URLs:  urls,
		Error: "",
	}, nil
}

// extractURLsFromHTML extracts YouTube URLs from HTML content
func (c *Client) extractURLsFromHTML(html string) []string {
	var urls []string

	// Method 1: Extract from iframe src attributes
	iframeRegex := regexp.MustCompile(`<iframe[^>]+src=["']([^"']*youtube\.com[^"']*)["']`)
	matches := iframeRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, c.normalizeYouTubeURL(match[1]))
		}
	}

	// Method 2: Extract from embed src attributes
	embedRegex := regexp.MustCompile(`<embed[^>]+src=["']([^"']*youtube\.com[^"']*)["']`)
	matches = embedRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, c.normalizeYouTubeURL(match[1]))
		}
	}

	// Method 3: Extract from data attributes
	dataRegex := regexp.MustCompile(`data-[^=]*=["']([^"']*youtube\.com[^"']*)["']`)
	matches = dataRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, c.normalizeYouTubeURL(match[1]))
		}
	}

	// Method 4: Extract direct YouTube watch URLs
	watchRegex := regexp.MustCompile(`https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`)
	matches = watchRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, fmt.Sprintf("https://www.youtube.com/watch?v=%s", match[1]))
		}
	}

	// Method 5: Extract from embed URLs
	embedURLRegex := regexp.MustCompile(`https?://(?:www\.)?youtube\.com/embed/([a-zA-Z0-9_-]{11})`)
	matches = embedURLRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, fmt.Sprintf("https://www.youtube.com/watch?v=%s", match[1]))
		}
	}

	return urls
}

// normalizeYouTubeURL converts various YouTube URL formats to standard watch URL
func (c *Client) normalizeYouTubeURL(url string) string {
	// Extract video ID from various formats
	videoIDRegex := regexp.MustCompile(`(?:v=|/)([a-zA-Z0-9_-]{11})`)
	matches := videoIDRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", matches[1])
	}

	// If already a watch URL, return as-is
	if strings.Contains(url, "youtube.com/watch") {
		return url
	}

	return ""
}

// removeDuplicates removes duplicate URLs from a slice
func (c *Client) removeDuplicates(urls []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, url := range urls {
		if url != "" && !seen[url] {
			seen[url] = true
			result = append(result, url)
		}
	}
	return result
}

// GetSessionCookies extracts session cookies from current browser context
func (c *Client) GetSessionCookies(ctx context.Context) ([]SessionCookie, error) {
	var cookies []SessionCookie
	var cookieList []*network.Cookie

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookieList, err = network.GetCookies().Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, err
	}

	for _, cookie := range cookieList {
		var expires int64
		if cookie.Expires > 0 {
			expires = int64(cookie.Expires)
		}
		cookies = append(cookies, SessionCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  expires,
			HTTPOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
			SameSite: string(cookie.SameSite),
		})
	}

	return cookies, nil
}

// SerializeCookies converts cookies to JSON string for storage
func SerializeCookies(cookies []SessionCookie) (string, error) {
	data, err := json.Marshal(cookies)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeCookies converts JSON string back to cookies
func DeserializeCookies(data string) ([]SessionCookie, error) {
	var cookies []SessionCookie
	err := json.Unmarshal([]byte(data), &cookies)
	return cookies, err
}

