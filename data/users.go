package data

import (
	"encoding/json"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"

	"github.com/goware/geotools"
	"github.com/goware/lg"

	"upper.io/bond"
	"upper.io/db.v3"
)

type User struct {
	ID        int64  `db:"id,pk,omitempty" json:"id" facebook:"-"`
	Username  string `db:"username" json:"username" facebook:"id,required"`
	Email     string `db:"email" json:"email" facebook:"email"`
	Name      string `db:"name" json:"name" facebook:"name"`
	AvatarURL string `db:"avatar_url" json:"avatarUrl"`

	// facebook related friends
	FirstName string `db:"-" json:"-" facebook:"first_name"`
	Friends   struct {
		Summary struct {
			TotalCount int32 `facebook:"total_count"`
		} `facebook:"summary"`
	} `db:"-" json:"-" facebook:"friends"`

	AccessToken string         `db:"access_token" json:"-"`
	DeviceToken *string        `db:"device_token,omitempty" json:"-"`
	Network     string         `db:"network" json:"network"`
	LoggedIn    bool           `db:"logged_in" json:"-"`
	IsAdmin     bool           `db:"is_admin" json:"isAdmin"`
	LastLogInAt *time.Time     `db:"last_login_at" json:"lastLoginAt"`
	Geo         geotools.Point `db:"geo" json:"-"`
	Etc         UserEtc        `db:"etc,jsonb" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

type UserEtc struct {
	// Store user's current neighbourhood whereabouts
	LocaleID      int64  `json:"localeId"`
	HasAgreedNDA  bool   `json:"hasAgreedNDA"`
	FirstName     string `json:"firstName"`
	FbFriendCount int32  `json:"fbFriendCount"`
}

type UserStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
} = &User{}

func (u *User) CollectionName() string {
	return `users`
}

func (u *User) BeforeCreate(bond.Session) error {
	u.Geo = *geotools.NewPointFromLatLng(0, 0) // set to zero location
	return nil
}

// SetLocation sets the user geo location
func (u *User) SetLocation(lat, lon float64) error {
	lg.Infof("user(%d) update loc (%f,%f)", u.ID, lat, lon)
	u.Geo = *geotools.NewPointFromLatLng(lat, lon)
	return DB.Save(u)
}

func (u *User) DistanceToPlaces(places ...*Place) {
	userLoc := geotools.LatLngFromPoint(u.Geo)
	for _, p := range places {
		pLoc := geotools.LatLngFromPoint(p.Geo)
		p.Distance = DistanceTo(userLoc, pLoc)
	}
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

// NewSessionUser returns a session user from jwt auth token
func NewSessionUser(tok string) (*User, error) {
	token, err := token.Decode(tok)
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
	return nil, api.ErrInvalidSession
}
