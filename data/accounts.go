package data

import (
	"encoding/json"
	"time"

	"upper.io/bond"
	"upper.io/db"
)

type Account struct {
	ID        int64  `db:"id,pk,omitempty" json:"id" facebook:"-"`
	Username  string `db:"username" json:"username" facebook:"id,required"`
	Email     string `db:"email" json:"email" facebook:"email"`
	Name      string `db:"name" json:"name" facebook:"name"`
	AvatarURL string `db:"avatar_url" json:"avatar_url"`

	AccessToken string     `db:"access_token" json:"-"`
	Network     string     `db:"network" json:"network"`
	LoggedIn    bool       `db:"logged_in" json:"logged_in"`
	LastLogInAt *time.Time `db:"last_login_at" json:"last_login_at"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// Authenticated user with jwt embed
type AuthUser struct {
	*Account
	JWT string `json:"jwt"`
}

type AccountStore struct {
	bond.Store
}

func (u *Account) CollectionName() string {
	return `accounts`
}

func (s AccountStore) FindByUsername(username string) (*Account, error) {
	return s.FindOne(db.Cond{"username": username})
}

func (s AccountStore) FindByID(ID int64) (*Account, error) {
	return s.FindOne(db.Cond{"id": ID})
}

func (s AccountStore) FindOne(cond db.Cond) (*Account, error) {
	var a *Account
	if err := s.Find(cond).One(&a); err != nil {
		return nil, err
	}
	return a, nil
}

// AuthUser wraps a user with JWT token
func NewAuthUser(user *Account) (*AuthUser, error) {
	token, err := GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &AuthUser{Account: user, JWT: token.Raw}, nil
}

// NewSessionUser returns a session user from jwt auth token
func NewSessionUser(tok string) (*Account, error) {
	token, err := DecodeToken(tok)
	if err != nil {
		return nil, err
	}

	userID, err := token.Claims["user_id"].(json.Number).Int64()
	if err != nil {
		return nil, err
	}

	// find a logged in user with the given id
	return DB.Account.FindOne(db.Cond{"id": userID, "logged_in": true})
}
