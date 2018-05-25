package data

import (
	"upper.io/bond"
)

type PriorityMerchant struct {
	ID int64 `db:"place_id" json:"place_id"`
}

type PriorityMerchantStore struct {
	bond.Store
}

func (p *PriorityMerchant) CollectionName() string {
	return `priority_merchants`
}
