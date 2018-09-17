package connect

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

func SetupReporter(conf ReporterConfig) {
	client, err := NewReporter(nil, conf)
	if err != nil {
		lg.Alertf("failed to startup reporter: %v", err)
	}
	RT = client
}

type ReporterClient struct {
	client *http.Client

	baseURL *url.URL
	Debug   bool // turn on debugging
}

var (
	RT *ReporterClient

	ErrReporterClient = errors.New("reporter is not initialized")
)

func NewReporter(httpClient *http.Client, conf ReporterConfig) (*ReporterClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u, err := url.Parse(conf.ServerURL)
	if err != nil {
		return nil, err
	}
	c := &ReporterClient{
		client:  httpClient,
		baseURL: u,
	}
	return c, nil
}

func GetTrendingIDs(params url.Values) ([]int64, error) {
	if RT == nil {
		return nil, ErrReporterClient
	}

	rel, err := url.Parse("/trend")
	if err != nil {
		return nil, err
	}
	u := RT.baseURL.ResolveReference(rel)
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := RT.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := struct {
		IDs []int64 `json:"IDs"`
	}{}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return result.IDs, nil
}
