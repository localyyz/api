package connect

import (
	"errors"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/connect/facebook"
	"bitbucket.org/pxue/api/lib/ws"
)

type Connect interface {
	ExchangeToken(code string) (string, error) // exchange short-lived for longlived token
	GetUser(u *data.Account) error             // query connecting network for user data
}

var (
	ErrUnknownSocial = errors.New("unknown social")
)

func NewSocialAuth(connectID string) (Connect, error) {
	switch connectID {
	case "facebook":
		return fbconnect.New(), nil
	default:
		return nil, ErrUnknownSocial
	}

	return nil, ws.ErrUnrechable
}
