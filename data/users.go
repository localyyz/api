package data

import (
	"encoding/json"
	"time"

	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/goware/geotools"

	"upper.io/bond"
	"upper.io/db"
)

type User struct {
	ID        int64  `db:"id,pk,omitempty" json:"id" facebook:"-"`
	Username  string `db:"username" json:"username" facebook:"id,required"`
	Email     string `db:"email" json:"email" facebook:"email"`
	Name      string `db:"name" json:"name" facebook:"name"`
	AvatarURL string `db:"avatar_url" json:"avatarUrl"`

	AccessToken string         `db:"access_token" json:"-"`
	Network     string         `db:"network" json:"network"`
	LoggedIn    bool           `db:"logged_in" json:"-"`
	LastLogInAt *time.Time     `db:"last_login_at" json:"lastLoginAt"`
	Geo         geotools.Point `db:"geo" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

type UserContext struct {
	Commented bool `json:"commented"`
	Liked     bool `json:"liked"`
}

// Authenticated user with jwt embed
type AuthUser struct {
	*User
	JWT string `json:"jwt"`
}

type LocateUser struct {
	*User
	Geo    geotools.Point `json:"geo"`
	Locale *Locale        `json:"locale"`
}

type UserStore struct {
	bond.Store
}

func (u *User) CollectionName() string {
	return `users`
}

// SetLocation sets the user geo location
func (u *User) SetLocation(lat, lon float64) error {
	u.Geo = *geotools.NewPointFromLatLng(lat, lon)
	return DB.Save(u)
}

func (s UserStore) FindByUsername(username string) (*User, error) {
	return s.FindOne(db.Cond{"username": username})
}

func (s UserStore) FindByID(ID int64) (*User, error) {
	return s.FindOne(db.Cond{"id": ID})
}

func (s UserStore) FindOne(cond db.Cond) (*User, error) {
	var a *User
	if err := s.Find(cond).One(&a); err != nil {
		return nil, err
	}
	return a, nil
}

// AuthUser wraps a user with JWT token
func NewAuthUser(user *User) (*AuthUser, error) {
	token, err := GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &AuthUser{User: user, JWT: token.Raw}, nil
}

// NewSessionUser returns a session user from jwt auth token
func NewSessionUser(tok string) (*User, error) {
	token, err := DecodeToken(tok)
	if err != nil {
		return nil, err
	}

	rawUserID, ok := token.Claims["user_id"].(json.Number)
	if ok {
		userID, err := rawUserID.Int64()
		if err != nil {
			return nil, err
		}

		// find a logged in user with the given id
		return DB.User.FindOne(db.Cond{"id": userID, "logged_in": true})
	}
	return nil, utils.ErrInvalidSession
}
