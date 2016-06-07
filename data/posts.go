package data

import (
	"fmt"
	"time"

	"bitbucket.org/pxue/api/lib/ws"

	"upper.io/bond"
	"upper.io/db"
)

type Post struct {
	ID         int64      `db:"id,pk,omitempty" json:"id"`
	UserID     int64      `db:"user_id" json:"id"`
	LocationID int64      `db:"location_id" json:"location_id"`
	Filter     PostFilter `db:"filter" json:"filter"`

	Caption  string
	ImageURL string // TODO imageID? for now, just serve the image is fine

	Likes    uint32 `db:"likes" json:"likes"`
	Comments uint32 `db:"comments" json:"comments"`
	Score    uint64 `db:"score" json:"-"` // internal score

	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"upated_at,omitempty"`
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
} = &Post{}

var (
	postFilters = []string{"none"}
)

func (p *Post) CollectionName() string {
	return `posts`
}

func (p *Post) BeforeCreate(bond.Session) error {
	p.CreatedAt = GetTimeUTCPointer()

	return nil
}

func (s *PostStore) GetTrending(cursor *ws.Page) ([]*Post, error) {
	q := s.Find().Sort("-score")
	if cursor != nil {
		q = cursor.UpdateQueryUpper(q)
	}
	var posts []*Post
	if err := q.All(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *PostStore) GetFresh(cursor *ws.Page) ([]*Post, error) {
	q := s.Find().Sort("-created_at")
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
}

// Update comment count on the post...
func (p *Post) UpdateCommentCount() {
	count, err := DB.Comment.Find(db.Cond{"post_id": p.ID}).Count()
	if err == nil {
		p.Comments = uint32(count)
		DB.Save(p)
	}
}

func (s *PostStore) FindByID(postID int64) (*Post, error) {
	return s.FindOne(db.Cond{"id": postID})
}

func (s *PostStore) FindOne(cond db.Cond) (*Post, error) {
	var p *Post
	if err := s.Find(cond).One(&p); err != nil {
		return nil, err
	}
	return p, nil
}

// String returns the string value of the status.
func (s PostFilter) String() string {
	return postFilters[s]
}

// MarshalText satisfies TextMarshaler
func (s PostFilter) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *PostFilter) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(postFilters); i++ {
		if enum == postFilters[i] {
			*s = PostFilter(i)
			return nil
		}
	}
	return fmt.Errorf("unknown post filter %s", enum)
}
