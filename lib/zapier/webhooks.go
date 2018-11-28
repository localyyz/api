package zapier

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type Webhook struct {
	hookURL string
}

func NewWebhook(url string) *Webhook {
	return &Webhook{url}
}

type ZapierPayload struct {
	Value  interface{} `json:"value"`
	SentAt time.Time   `json:"sentAt"`
}

func (hk *Webhook) Post(v interface{}) error {
	payload := ZapierPayload{
		Value:  v,
		SentAt: time.Now().UTC(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(hk.hookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(t))
	}

	return nil
}

func (hk *Webhook) Put(v interface{}) error {
	payload := ZapierPayload{
		Value:  v,
		SentAt: time.Now().UTC(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", hk.hookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(t))
	}

	return nil
}
