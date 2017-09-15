package stripe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Referenced + Modified
// https://github.com/stripe/stripe-go

type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	BaseURL       *url.URL
	Debug         bool // turn on debugging
	StripeAccount string

	// User agent used when communicating with the Stripe API.
	UserAgent string

	common service

	Token *TokenService
}

type service struct {
	client *Client
}

const (
	defaultBaseURL     = "https://api.stripe.com/"
	defaultApiVer      = "v1/"
	defaultContentType = "application/x-www-form-urlencoded"

	userAgent           = `go-stripe`
	clientAccountHeader = `Stripe-Account`
)

func NewClient(accountID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:        httpClient,
		UserAgent:     userAgent,
		BaseURL:       baseURL,
		StripeAccount: accountID,
	}
	c.common.client = c

	c.Token = (*TokenService)(&c.common)

	return c
}

func (c *Client) NewRequest(method, urlStr string, form *RequestValues) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(defaultApiVer + urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.Reader
	if form != nil && !form.Empty() {
		buf = bytes.NewBufferString(form.Encode())
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	// Support the value of the old Account field for now.
	if accountID := strings.TrimSpace(c.StripeAccount); accountID != "" {
		req.Header.Set("Stripe-Account", accountID)
	}

	if form != nil {
		req.Header.Set("Content-Type", defaultContentType)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
// TODO: Rate limiting
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		//if e, ok := err.(*url.Error); ok {
		//if url, err := url.Parse(e.URL); err == nil {
		//e.URL = sanitizeURL(url).String()
		//return nil, e
		//}
		//}

		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			if c.Debug {
				b, _ := ioutil.ReadAll(resp.Body)
				fmt.Printf("[stripe]: %s\n", string(b))
				err = json.Unmarshal(b, v)
			} else {
				err = json.NewDecoder(resp.Body).Decode(v)
			}

			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}
