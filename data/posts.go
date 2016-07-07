package data

import (
	"fmt"
	"time"

	"github.com/goware/lg"

	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"upper.io/bond"
	"upper.io/db"
)

type Post struct {
	ID      int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	UserID  int64      `db:"user_id" json:"userId"`
	PlaceID int64      `db:"place_id" json:"placeId"`
	Filter  PostFilter `db:"filter" json:"filter"`

	Caption  string `db:"caption" json:"caption"`
	ImageURL string `db:"image_url" json:"imageUrl"`

	Likes    uint32 `db:"likes" json:"likes"`
	Comments uint32 `db:"comments" json:"comments"`
	Score    uint64 `db:"score" json:"-"` // internal score for trending
	Featured int64  `db:"featured" json:"featured"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type PostPresenter struct {
	*Post
	User    *User        `json:"user"`
	Place   *Place       `json:"place"`
	Context *UserContext `json:"context"`
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
	postFilters = []string{"none"}
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

func (store *PostStore) GetFresh(cursor *ws.Page, cond db.Cond) ([]*Post, error) {
	q := store.Find().Sort("-created_at")
	if len(cond) > 0 {
		q = q.Where(cond) // filter by first
	}

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
	if err := DB.Save(p); err != nil {
		lg.Errorf("failed to update post score: %v", err)
	}
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
	isLimited, err := DB.UserPoint.IsLimited(p.UserID)
	if err != nil || isLimited {
		return err
	}

	up := &UserPoint{UserID: p.UserID, PostID: p.ID, PlaceID: p.PlaceID}
	if err := DB.UserPoint.Save(up); err != nil {
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
