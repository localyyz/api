package data

import (
	"fmt"
	"time"

	"github.com/goware/geotools"
	"github.com/goware/lg"

	"upper.io/bond"
	"upper.io/db.v3"
)

type User struct {
	ID          int64       `db:"id,pk,omitempty" json:"id" facebook:"-"`
	Username    string      `db:"username" json:"username" facebook:"id,required"`
	Email       string      `db:"email" json:"email" facebook:"email"`
	EmailStatus EmailStatus `db:"email_status" json:"emailStatus"`
	Name        string      `db:"name" json:"name" facebook:"name"`
	AvatarURL   string      `db:"avatar_url" json:"avatarUrl"`

	// facebook related fields
	FirstName string `db:"-" json:"-" facebook:"first_name"`
	Gender    string `db:"-" json:"-" facebook:"gender"`

	AccessToken  string         `db:"access_token" json:"-"`
	PasswordHash string         `db:"password_hash,omitempty" json:"-"`
	DeviceToken  *string        `db:"device_token,omitempty" json:"-"`
	InviteCode   string         `db:"invite_code" json:"inviteCode"` // Auto generated invite hash
	Network      string         `db:"network" json:"network"`
	LoggedIn     bool           `db:"logged_in" json:"-"`
	IsAdmin      bool           `db:"is_admin" json:"isAdmin"`
	LastLogInAt  *time.Time     `db:"last_login_at" json:"lastLoginAt"`
	Geo          geotools.Point `db:"geo" json:"-"`
	Etc          UserEtc        `db:"etc,jsonb" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

type EmailStatus uint

const (
	EmailStatusUnknown EmailStatus = iota
	EmailStatusUnconfirmed
	EmailStatusConfirmed
)

type UserGender uint

const (
	UserGenderUnknown UserGender = iota
	UserGenderMale
	UserGenderFemale
)

type UserEtc struct {
	// Store user's current neighbourhood whereabouts
	LocaleID         int64      `json:"localeId"`
	FirstName        string     `json:"firstName"`
	Gender           UserGender `json:"gender"`
	InvitedBy        int64      `json:"invitedBy"`
	StripeCustomerID string     `json:"stripeCustomerId"`
}

type UserStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &User{}

var (
	userGenders   = []string{"unknown", "male", "female"}
	emailStatuses = []string{"unknown", "unconfirmed", "confirmed"}
)

func (u *User) CollectionName() string {
	return `users`
}

func (u *User) BeforeCreate(bond.Session) error {
	u.Geo = *geotools.NewPointFromLatLng(0, 0) // set to zero location
	u.InviteCode = RandString(5)               // random user invite_code hash
	//TODO: unlikely event of conflict, do something

	return nil
}

func (u *User) BeforeUpdate(bond.Session) error {
	u.UpdatedAt = GetTimeUTCPointer()

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

func (s UserStore) FindByInviteCode(code string) (*User, error) {
	return s.FindOne(db.Cond{"invite_code": code})
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

// String returns the string value of the gender.
func (s UserGender) String() string {
	return userGenders[s]
}

// MarshalText satisfies TextMarshaler
func (s UserGender) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *UserGender) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(userGenders); i++ {
		if enum == userGenders[i] {
			*s = UserGender(i)
			return nil
		}
	}
	return fmt.Errorf("unknown user gender %s", enum)
}

// String returns the string value of the status.
func (s EmailStatus) String() string {
	return emailStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s EmailStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *EmailStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(emailStatuses); i++ {
		if enum == emailStatuses[i] {
			*s = EmailStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown email status %s", enum)
}
