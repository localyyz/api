package data

import (
	"upper.io/bond"
)

type PriorityMerchant struct {
	ID int64 `db:"id" json:"id"`
}

type PriorityMerchantStore struct {
	bond.Store
}

func (p *PriorityMerchant) CollectionName() string {
	return `priority_merchants`
}
