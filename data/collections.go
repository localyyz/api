package data

import (
	"time"

	"fmt"
	"github.com/pressly/lg"
	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Collection struct {
	ID          int64         `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
	ImageURL    string        `db:"image_url" json:"imageUrl"`
	Gender      ProductGender `db:"gender" json:"gender"`

	PlaceIDs   *postgresql.Int64Array  `db:"place_ids" json:"-"`
	Categories *postgresql.StringArray `db:"categories" json:"-"`

	Ordering  int32      `db:"ordering" json:"ordering"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`

	Lightning bool             `db:"lightning" json:"lightning"`
	StartTime *time.Time       `db:"time_start" json:"startTime"`
	EndTime   *time.Time       `db:"time_end" json:"endTime"`
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
	ID           int64      `db:"id,pk,omitempty" json:"id,omitempty"`
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
	CollectionStatusActive                           //1
	CollectionStatusInactive                         //2
	CollectionStatusQueued                           //3
)

var collectionStatuses = []string{"active", "inactive", "queued"}

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

/*
	Returns the completion percent(0.0-1) of a collection
*/
func (store CollectionStore) GetCollectionCheckouts(collection *Collection) int {
	var checkouts []*Checkout
	DB.Select("ck.*").
		From("collection_products as cp").
		Join("cart_items as ci").On("cp.product_id = ci.product_id").
		Join("checkouts as ck").On("ci.checkout_id = ck.id").
		Where(
			db.And(
				db.Cond{
					"cp.collection_id": collection.ID,
					"ck.status":        CheckoutStatusPaymentSuccess,
				},
				db.Raw("ci.checkout_id IS NOT NULL"),
			),
		).All(&checkouts)

	return len(checkouts)

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

func CheckCollectionExpireTime() {
	collections, err := DB.Collection.FindAll(db.Cond{"lightning": true, "status": CollectionStatusActive})
	if err != nil {
		return
	}
	for _, collection := range collections {
		time := time.Now()
		if collection.EndTime != nil {
			if collection.EndTime.Before(time) {
				collection.Status = CollectionStatusInactive
				if err := DB.Save(collection); err != nil {
					lg.Warn("Error: failed to set the collection status to inactive")
				}
			}
		}
	}
}

func CheckCapLimit() {
	collections, err := DB.Collection.FindAll(db.Cond{"lightning": true, "status": CollectionStatusActive})
	if err != nil {
		lg.Warn("Error: could not retrieve collections for cron job")
	}
	for _, collection := range collections {
		totalCheckouts := DB.Collection.GetCollectionCheckouts(collection)
		if int64(totalCheckouts) >= collection.Cap {
			collection.Status = CollectionStatusInactive
			if err := DB.Save(collection); err != nil {
				lg.Warn("Error: failed to set the collection status to inactive")
			}
		}
	}

}
