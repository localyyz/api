package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type UserAccessStore struct {
	bond.Store
}

func (store UserAccessStore) EditorAccess(userID int64) ([]*UserAccess, error) {
	cond := db.And(
		db.Cond{
			"user_id": userID,
		},
		db.Or(
			// user either admin, or promoter
			db.Cond{"admin": true},
			db.Cond{"promoter": true},
		),
	)
	var access []*UserAccess
	if err := store.Find(cond).All(&access); err != nil {
		return nil, err
	}

	return access, nil
}

func (store UserAccessStore) FindAll(cond db.Cond) ([]*UserAccess, error) {
	var uas []*UserAccess
	if err := store.Find(cond).All(&uas); err != nil {
		return nil, err
	}
	return uas, nil
}
