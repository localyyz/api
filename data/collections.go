package data

import (
	"time"

	"fmt"

	"github.com/pkg/errors"
	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Collection struct {
	ID          int64         `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
	ImageURL    string        `db:"image_url" json:"imageUrl"`
	ImageWidth  int64         `db:"image_width" json:"imageWidth"`
	ImageHeight int64         `db:"image_height" json:"imageHeight"`
	Gender      ProductGender `db:"gender" json:"gender"`
	Featured    bool          `db:"featured" json:"featured"`

	PlaceIDs   *postgresql.Int64Array  `db:"place_ids" json:"-"`
	Categories *postgresql.StringArray `db:"categories" json:"-"`

	Ordering  int32      `db:"ordering" json:"ordering"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`

	Lightning bool             `db:"lightning" json:"lightning"`
	StartAt   *time.Time       `db:"start_at" json:"startAt"`
	EndAt     *time.Time       `db:"end_at" json:"endAt"`
	Status    CollectionStatus `db:"status" json:"status"`
	Cap       int64            `db:"cap" json:"cap"`
}

type CollectionStore struct {
	bond.Store
}

func (*Collection) CollectionName() string {
	return `collections`
}

type CollectionProduct struct {
	CollectionID int64      `db:"collection_id" json:"collection_id"`
	ProductID    int64      `db:"product_id" json:"product_id"`
	CreatedAt    *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type CollectionProductStore struct {
	bond.Store
}

type CollectionStatus int32

const (
	_                        CollectionStatus = iota //0
	CollectionStatusQueued                           //1
	CollectionStatusActive                           //2
	CollectionStatusInactive                         //3
)

var collectionStatuses = []string{"-", "queued", "active", "inactive"}

func (*CollectionProduct) CollectionName() string {
	return `collection_products`
}

func (store CollectionStore) FindByID(ID int64) (*Collection, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store CollectionStore) FindOne(cond db.Cond) (*Collection, error) {
	var collection *Collection
	err := DB.Collection.Find(cond).One(&collection)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (store CollectionStore) FindAll(cond db.Cond) ([]*Collection, error) {
	var collections []*Collection
	err := DB.Collection.Find(cond).All(&collections)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (store CollectionProductStore) FindByCollectionID(collectionID int64) ([]*CollectionProduct, error) {
	return DB.CollectionProduct.FindAll(db.Cond{"collection_id": collectionID})
}

func (store CollectionProductStore) FindAll(cond db.Cond) ([]*CollectionProduct, error) {
	var collections []*CollectionProduct
	err := DB.CollectionProduct.Find(cond).All(&collections)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (c CollectionStatus) String() string {
	return collectionStatuses[c]
}

func (c CollectionStatus) MarshallText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *CollectionStatus) UnmarshallText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(collectionStatuses); i++ {
		if enum == collectionStatuses[i] {
			*c = CollectionStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown collection status %s", enum)
}

/*
	Returns the total number of successfull checkouts of a collection
*/
func (c *Collection) GetCheckoutCount() (int, error) {
	row, err := DB.Select(db.Raw("count(1) as _t")).
		From("collection_products as cp").
		LeftJoin("cart_items as ci").On("cp.product_id = ci.product_id").
		LeftJoin("carts c").On("c.id = ci.cart_id").
		Where(
			db.Cond{
				"cp.collection_id": c.ID,
				"c.status":         CartStatusPaymentSuccess,
			},
		).QueryRow()
	if err != nil {
		return 0, errors.Wrap(err, "collection checkout prepare")
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "collection checkout scan")
	}

	return count, nil
}

// find and de-activate active collections that has expired
func UpdateCollectionStatus() {
	// expire collections
	DB.Exec(`UPDATE collections SET status = 3 WHERE lightning = true AND NOW() at time zone 'utc' > end_at and status = 2`)
	// activate collections
	DB.Exec(`UPDATE collections SET status = 2 WHERE lightning = true AND NOW() at time zone 'utc' > start_at and status = 1`)
}
