package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Client struct {
	client *http.Client

	headers map[string]string
	cookies map[string]*http.Cookie

	BaseURL *url.URL
	Debug   bool

	common service

	Deal        *DealService
	Cart        *CartService
	ExpressCart *ExpressCartService
	User        *UserService
	UserColl    *UserCollService
}

type service struct {
	client *Client
}

func NewClient(apiURL string) (*Client, error) {
	url, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		BaseURL: url,
		client:  http.DefaultClient,
		headers: map[string]string{},
		cookies: map[string]*http.Cookie{},
	}
	c.common.client = c
	c.Cart = (*CartService)(&c.common)
	c.ExpressCart = (*ExpressCartService)(&c.common)
	c.User = (*UserService)(&c.common)
	c.Deal = (*DealService)(&c.common)
	c.User = (*UserService)(&c.common)
	c.UserColl = (*UserCollService)(&c.common)
	return c, nil
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	defer func() {
		if c.Debug {
			b, _ := httputil.DumpRequest(req, true)
			fmt.Printf("[apiclient] %s", string(b))
		}
	}()
	if err != nil {
		return nil, err
	}

	for key, value := range c.headers {
		req.Header.Add(key, value)
	}
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	return req, nil
}

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

		return resp, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	var apiError error
	switch resp.StatusCode {
	case 200, 201, 204, 301, 302:
		// do nothing, valid response.
	default:
		// parse the error and return wrapped error type
		apiError = NewError(resp)
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			if c.Debug {
				b, _ := httputil.DumpResponse(resp, true)
				fmt.Printf("[apiclient]: %s\n", string(b))
			}
			err = json.NewDecoder(resp.Body).Decode(v)
			// ignore this error if apiError is already set
			if err != nil && err != io.EOF && apiError == nil {
				return resp, err
			}
		}
	}

	return resp, apiError
}

func (c *Client) AddHeader(key string, value string) {
	c.headers[key] = value
}

func (c *Client) JWT(value string) {
	c.AddHeader("Authorization", "BEARER "+value)
}
