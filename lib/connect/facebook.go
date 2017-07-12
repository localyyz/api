package connect

import (
	"errors"
	"fmt"

	"bitbucket.org/moodie-app/moodie-api/data"

	"upper.io/db.v3"

	"github.com/goware/lg"
	fb "github.com/huandu/facebook"
)

var (
	FB *Facebook
	// Facebook API version
	// See: https://developers.facebook.com/docs/apps/changelog for updates
	FBVersion = "v2.9"
)

type Facebook struct {
	*fb.App
}

func SetupFacebook(conf Config) {
	FB = &Facebook{
		App: &fb.App{
			AppId:     conf.AppId,
			AppSecret: conf.AppSecret,
		},
	}
}

func (f *Facebook) Login(token, inviteCode string) (*data.User, error) {
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
		user = &data.User{Network: `facebook`}
	}

	if token != "" {
		// first time login, or new login, exchange for long-lived token
		newToken, _, err := f.ExchangeToken(token)
		if err != nil {
			return nil, err
		}
		user.AccessToken = newToken
	}

	// always refetch facebook detail on login
	if err := f.GetUser(user); err != nil {
		return nil, err
	}

	user.LoggedIn = true
	user.LastLogInAt = data.GetTimeUTCPointer()

	if inviteCode != "" {
		invitor, err := data.DB.User.FindByInviteCode(inviteCode)
		if err != nil {
			lg.Alertf("invitor with code %s lookup error: %v", inviteCode, err)
		}
		lg.Info("new user invited by: %d", invitor.ID)
		user.Etc = data.UserEtc{
			InvitedBy: invitor.ID,
		}
		// TODO return invalid code error?
	}

	if err := data.DB.User.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (f *Facebook) GetUser(u *data.User) error {
	if u.AccessToken == "" {
		return errors.New("invalid token")
	}

	var resp fb.Result
	var err error
	params := fb.Params{
		"fields":       "id,first_name,name,email,timezone,link,friends",
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
	lg.Warnf("%+v", u.Friends)
	u.AvatarURL = fmt.Sprintf("https://graph.facebook.com/%s/picture?type=large", u.Username)
	u.Etc.FirstName = u.FirstName
	u.Etc.FbFriendCount = u.Friends.Summary.TotalCount

	return nil
}
