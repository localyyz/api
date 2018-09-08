package xchange

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

type XChange struct {
	Rates map[string]float64
}

var (
	ErrUnknownCurrency = errors.New("unknown currency")
	instance           *XChange

	tokenRgx   = regexp.MustCompile(`"\w+":\d+.\d+`)
	tokenSplit = []byte(":")
)

func New() (*XChange, error) {
	instance = &XChange{Rates: map[string]float64{}}
	if err := instance.LoadRates(); err != nil {
		return nil, err
	}
	return instance, nil
}

func (x *XChange) LoadRates() error {
	// load up from shopify
	resp, err := http.Get("https://cdn.shopify.com/s/javascripts/currencies.js")
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rates := tokenRgx.FindAll(raw, -1)
	for _, rate := range rates {
		tokens := bytes.Split(rate, tokenSplit)

		key := string(bytes.Trim(tokens[0], `"`))
		val, err := strconv.ParseFloat(string(tokens[1]), 64)
		if err != nil {
			return err
		}
		x.Rates[key] = val
	}

	return nil
}

func (x *XChange) Convert(amount float64, from, to string) float64 {
	return (amount * x.Rates[from]) / x.Rates[to]
}

func Convert(amount float64, from, to string) float64 {
	if instance == nil {
		return amount
	}
	return instance.Convert(amount, from, to)
}

func ToUSD(amount float64, from string) float64 {
	return Convert(amount, from, "USD")
}
