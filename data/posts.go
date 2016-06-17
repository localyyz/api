package data

import (
	"fmt"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"upper.io/bond"
	"upper.io/db"
)

type Post struct {
	ID         int64      `db:"id,pk,omitempty" json:"id"`
	UserID     int64      `db:"user_id" json:"userId"`
	LocationID int64      `db:"location_id" json:"locationId"`
	Filter     PostFilter `db:"filter" json:"filter"`

	Caption  string
	ImageURL string // TODO imageID? for now, just serve the image is fine

	Likes    uint32 `db:"likes" json:"likes"`
	Comments uint32 `db:"comments" json:"comments"`
	Score    uint64 `db:"score" json:"-"` // internal score for trending
	Featured int64  `db:"featured" json:"featured"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type PostStore struct {
	bond.Store
}

type PostFilter uint

const (
	FilterNone PostFilter = iota
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasAfterCreate
} = &Post{}

var (
	postFilters     = []string{"none"}
	DailyPointLimit = 3
)

func (p *Post) CollectionName() string {
	return `posts`
}

func (store *PostStore) GetTrending(cursor *ws.Page) ([]*Post, error) {
	q := store.Find().Sort("-score")
	if cursor != nil {
		q = cursor.UpdateQueryUpper(q)
	}
	var posts []*Post
	if err := q.All(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (store *PostStore) GetFresh(cursor *ws.Page) ([]*Post, error) {
	q := store.Find().Sort("-created_at")
	if cursor != nil {
		q = cursor.UpdateQueryUpper(q)
	}
	var posts []*Post
	if err := q.All(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// Update likes count on the post...
func (p *Post) UpdateLikeCount() {
	count, err := DB.Like.Find(db.Cond{"post_id": p.ID}).Count()
	if err == nil {
		p.Likes = uint32(count)
		DB.Save(p)
	}
	p.UpdateScore()
}

// Update comment count on the post...
func (p *Post) UpdateCommentCount() {
	count, err := DB.Comment.Find(db.Cond{"post_id": p.ID}).Count()
	if err == nil {
		p.Comments = uint32(count)
		DB.Save(p)
	}
	p.UpdateScore()
}

// Update score
func (p *Post) UpdateScore() {
	p.Score = uint64(p.CreatedAt.Unix()) + uint64(p.Likes) + uint64(p.Comments)
	DB.Save(p)
}

func (p *Post) BeforeCreate(bond.Session) error {
	p.CreatedAt = GetTimeUTCPointer()

	// score is current time in utc.
	// NOTE: if no one likes or comments. trending = fresh
	p.Score = uint64(p.CreatedAt.Unix())

	return nil
}

func (p *Post) AfterCreate(sess bond.Session) error {
	// add to user points
	// TODO: smarter throttling. ie 2 points max per venue, up to 3 venue..
	count, err := DB.UserPoint.CountByUserID(p.UserID)
	if err != nil {
		return err
	}
	if int(count) > DailyPointLimit { // do nothing
		return nil
	}

	if err := DB.UserPoint.Save(&UserPoint{UserID: p.UserID, PostID: p.ID}); err != nil {
		sess.Rollback()
		return err
	}
	return nil
}

func (store *PostStore) FindUserRecent(userID int64, optCursor ...*ws.Page) ([]*Post, error) {
	var posts []*Post
	q := store.Find(db.Cond{"user_id": userID}).Sort("-created_at")
	if err := q.All(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (store *PostStore) FindByID(postID int64) (*Post, error) {
	return store.FindOne(db.Cond{"id": postID})
}

func (store *PostStore) FindOne(cond db.Cond) (*Post, error) {
	var p *Post
	if err := store.Find(cond).One(&p); err != nil {
		return nil, err
	}
	return p, nil
}

func (store *PostStore) FindAll(cond db.Cond) ([]*Post, error) {
	var posts []*Post
	if err := store.Find(cond).All(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// String returns the string value of the status.
func (pf PostFilter) String() string {
	return postFilters[pf]
}

// MarshalText satisfies TextMarshaler
func (pf PostFilter) MarshalText() ([]byte, error) {
	return []byte(pf.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (pf *PostFilter) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(postFilters); i++ {
		if enum == postFilters[i] {
			*pf = PostFilter(i)
			return nil
		}
	}
	return fmt.Errorf("unknown post filter %s", enum)
}
