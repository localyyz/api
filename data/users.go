package data

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
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

	AccessToken  string     `db:"access_token" json:"-"`
	PasswordHash string     `db:"password_hash,omitempty" json:"-"`
	DeviceToken  *string    `db:"device_token,omitempty" json:"-"`
	InviteCode   string     `db:"invite_code" json:"inviteCode"` // Auto generated invite hash
	Network      string     `db:"network" json:"network"`
	LoggedIn     bool       `db:"logged_in" json:"-"`
	LastLogInAt  *time.Time `db:"last_login_at" json:"lastLoginAt"`
	Etc          UserEtc    `db:"etc" json:"etc"`

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
	FirstName string     `json:"firstName"`
	Gender    UserGender `json:"gender"`
	InvitedBy int64      `json:"invitedBy"`

	*postgresql.JSONBConverter
}

type UserStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &User{}
var _ sqlbuilder.ValueWrapper = &UserEtc{}

var (
	userGenders   = []string{"unknown", "male", "female"}
	emailStatuses = []string{"unknown", "unconfirmed", "confirmed"}
)

func (u *User) CollectionName() string {
	return `users`
}

func (u *User) BeforeCreate(bond.Session) error {
	u.InviteCode = RandString(5) // random user invite_code hash
	//TODO: unlikely event of conflict, do something

	return nil
}

func (u *User) BeforeUpdate(bond.Session) error {
	u.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (u *User) GetTotalCheckout(productID int64) (int, error) {
	row, err := DB.Select(db.Raw("count(1) as _t")).
		From("cart_items as ci").
		LeftJoin("carts as c").On("ci.cart_id = c.id").
		Where(
			db.Cond{
				"ci.product_id": productID,
				"c.user_id":     u.ID,
				"c.status":      CartStatusPaymentSuccess,
			},
		).QueryRow()
	if err != nil {
		return 0, errors.Wrap(err, "product checkout prepare")
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "product checkout scan")
	}

	return count, nil
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
