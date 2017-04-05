package store

import (
	"net/http"
	"strings"

	"github.com/goware/lg"
)

func (s *Store) CheckIsShopify() (bool, error) {
	if s.Website == "" {
		return false, nil
	}

	resp, err := http.Head(s.Website)
	if err != nil {
		return false, err
	}

	for _, v := range resp.Header {
		if !strings.Contains(v[0], "Shopify") {
			continue
		}

		lg.Warnf("%s,%s", s.Name, s.Website)
		return true, nil
	}

	return false, nil
}
