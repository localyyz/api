package stash

import "github.com/gomodule/redigo/redis"

func GetUserCollProdCount(collectionID int64) (int64, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}

	return DefaultClient.GetUserCollProdCount(collectionID)
}

func GetUserCollSavings(collectionID int64) (float64, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}

	return DefaultClient.GetUserCollSavings(collectionID)
}

func IncrUserCollProdCount(collectionID int64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}

	return DefaultClient.IncrUserCollProdCount(collectionID)
}

func IncrUserCollSavings(collectionID int64, savings float64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}

	return DefaultClient.IncrUserCollSavings(collectionID, savings)
}

func DecrUserCollProdCount(collectionID int64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}

	return DefaultClient.DecrUserCollProdCount(collectionID)
}

func DecrUserCollSavings(collectionID int64, savings float64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}

	return DefaultClient.DecrUserCollSavings(collectionID, savings)
}

func DelUserColl(collectionID int64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}

	return DefaultClient.DelUserColl(collectionID)
}

func (s *Stash) GetUserCollProdCount(collectionID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("ZSCORE", "usercollection:productcount", collectionID))
}

func (s *Stash) GetUserCollSavings(collectionID int64) (float64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Float64(conn.Do("ZSCORE", "usercollection:savings", collectionID))
}

func (s *Stash) IncrUserCollProdCount(collectionID int64) error {
	conn := s.conn()
	defer conn.Close()

	conn.Do("ZINCRBY", "usercollection:productcount", 1, collectionID)

	return nil
}

func (s *Stash) IncrUserCollSavings(collectionID int64, savings float64) error {
	conn := s.conn()
	defer conn.Close()

	conn.Do("ZINCRBY", "usercollection:savings", savings, collectionID)

	return nil
}

func (s *Stash) DecrUserCollProdCount(collectionID int64) error {
	conn := s.conn()
	defer conn.Close()

	conn.Do("ZINCRBY", "usercollection:productcount", -1, collectionID)

	return nil
}

func (s *Stash) DecrUserCollSavings(collectionID int64, savings float64) error {
	conn := s.conn()
	defer conn.Close()

	conn.Do("ZINCRBY", "usercollection:savings", -1*savings, collectionID)

	return nil
}

func (s *Stash) DelUserColl(collectionID int64) error {
	conn := s.conn()
	defer conn.Close()

	conn.Do("DEL", "usercollection:productcount", collectionID)
	conn.Do("DEL", "usercollection:savings", collectionID)

	return nil
}
