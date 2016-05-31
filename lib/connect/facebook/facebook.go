package fbconnect

import (
	"errors"

	"bitbucket.org/pxue/api/data"
	"github.com/goware/lg"
	fb "github.com/huandu/facebook"
)

var (
	AppId         string
	AppSecret     string
	OAuthCallback string
)

type FB struct {
	*fb.App
}

func New() *FB {
	return &FB{
		App: &fb.App{
			AppId:     AppId,
			AppSecret: AppSecret,
		},
	}
}

func (f *FB) ExchangeToken(code string) (string, error) {
	token, _, err := f.App.ExchangeToken(code)
	return token, err
}

func (f *FB) GetUser(u *data.Account) error {
	if u.AccessToken == "" {
		return errors.New("invalid token")
	}

	var resp fb.Result
	var err error
	params := fb.Params{
		"fields":       "id,name,email,picture.type(large),timezone,link",
		"access_token": u.AccessToken,
	}
	resp, err = fb.Get(`me`, params)
	if err != nil {
		lg.Warnf("fb getUser error: %v", err)
		return err
	}

	if err := resp.Decode(u); err != nil {
		lg.Warnf("fb decode error: %v", err)
		return err
	}

	// picture -> avatarurl
	if url := resp.GetField("picture", "data", "url"); url != nil {
		u.AvatarURL = url.(string)
	}

	return nil
}
