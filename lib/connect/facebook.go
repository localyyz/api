package connect

import (
	"errors"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"

	"upper.io/db.v2"

	"github.com/goware/lg"
	fb "github.com/huandu/facebook"
)

var (
	FB *FBConnect
	// Facebook API version
	// See: https://developers.facebook.com/docs/apps/changelog for updates
	FBVersion = "v2.6"
)

type FBConnect struct {
	*fb.App
}

func SetupFB(conf *Config) {
	FB = &FBConnect{
		App: &fb.App{
			AppId:     conf.AppId,
			AppSecret: conf.AppSecret,
		},
	}
}

func (f *FBConnect) Login(token string) (*data.User, error) {
	sess := f.App.Session(token)

	userID, err := sess.User()
	if err != nil {
		if e, ok := err.(*fb.Error); ok {
			if e.Code == 190 { // token expired
				return nil, ErrTokenExpired
			}
		}
		return nil, err
	}

	user, err := data.DB.User.FindByUsername(userID)
	if err != nil {
		if err != db.ErrNoMoreRows {
			return nil, err
		}
	}

	if user == nil {
		user = &data.User{
			AccessToken: token,
			Network:     `facebook`,
		}
		if err := f.GetUser(user); err != nil {
			return nil, err
		}

		// first time login, exchange for long-lived token
		token, _, err := FB.ExchangeToken(token)
		if err != nil {
			return nil, err
		}
		user.AccessToken = token
	}

	t := time.Now()
	user.LoggedIn = true
	user.LastLogInAt = &t

	if err := data.DB.User.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (f *FBConnect) GetUser(u *data.User) error {
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
