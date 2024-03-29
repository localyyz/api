package shopify

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
	client *http.Client // HTTP client used to communicate with the API.

	BaseURL *url.URL
	Debug   bool // turn on debugging

	// User agent used when communicating with the Shopify API.
	UserAgent string

	// NOTE authentication should be handled by external lib
	// just like the way github api was intended
	token string // Access token

	common service

	Billing          *BillingService
	Product          *ProductService
	Webhook          *WebhookService
	Shop             *ShopService
	Checkout         *CheckoutService
	ProductList      *ProductListService
	CollectionList   *CollectionListService
	Variant          *VariantService
	CustomCollection *CustomCollectionService
	PriceRule        *PriceRuleService
	DiscountCode     *DiscountCodeService
	Policy           *PolicyService
	ShippingZone     *ShippingZoneService
}

type service struct {
	client *Client
}

const (
	userAgent = `go-shopify`

	authHeader = `X-Shopify-Access-Token`
)

func NewClient(httpClient *http.Client, token string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{client: httpClient, UserAgent: userAgent, token: token}
	c.common.client = c

	c.Billing = (*BillingService)(&c.common)
	c.Product = (*ProductService)(&c.common)
	c.Webhook = (*WebhookService)(&c.common)
	c.Shop = (*ShopService)(&c.common)
	c.Checkout = (*CheckoutService)(&c.common)
	c.ProductList = (*ProductListService)(&c.common)
	c.CollectionList = (*CollectionListService)(&c.common)
	c.Variant = (*VariantService)(&c.common)
	c.CustomCollection = (*CustomCollectionService)(&c.common)
	c.PriceRule = (*PriceRuleService)(&c.common)
	c.DiscountCode = (*DiscountCodeService)(&c.common)
	c.Policy = (*PolicyService)(&c.common)
	c.ShippingZone = (*ShippingZoneService)(&c.common)
	return c
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
			fmt.Printf("[shopify] %s", string(b))
		}
	}()
	if err != nil {
		return nil, err
	}

	// Should have been done by the token transport, but
	// shopify doesn't use Autorization header, instead
	// uses this. so stupid
	req.Header.Set(authHeader, c.token)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
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

	// check for error response
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
				b, _ := httputil.DumpResponse(resp, true)
				fmt.Printf("[shopify]: %s\n", string(b))
			}
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}
