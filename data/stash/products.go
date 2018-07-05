package stash

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	Day               = 24 * time.Hour
	ProductViewExpire = 30 * Day
	ProductLiveExpire = 15 * time.Minute
)

func GetProductViews(productID int64) (int64, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}
	return DefaultClient.GetProductViews(productID)
}

func GetProductLiveViews(productID int64) (int64, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}
	return DefaultClient.GetProductLiveViews(productID)
}

func GetProductPurchases(productID int64) (int64, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}
	return DefaultClient.GetProductPurchases(productID)
}

func IncrProductViews(productID, userID int64) error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}
	return DefaultClient.IncrProductViews(productID, userID)
}

func IncrProductPurchases(productID int64) (int, error) {
	if DefaultClient == nil {
		return 0, ErrDefaultNotConfigured
	}
	return DefaultClient.IncrProductPurchases(productID)
}

func (s *Stash) GetProductViews(productID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("GET", s.productDataKey(productID, "views")))
}

func (s *Stash) GetProductLiveViews(productID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("ZCARD", s.productDataKey(productID, "live")))
}

func (s *Stash) GetProductPurchases(productID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("GET", s.productDataKey(productID, "buys")))
}

func (s *Stash) IncrProductViews(productID, userID int64) error {
	conn := s.conn()
	defer conn.Close()

	{ // view count
		k := s.productDataKey(productID, "views")
		conn.Send("MULTI")
		conn.Send("INCR", k)
		conn.Send("EXPIREAT", k, time.Now().Add(ProductViewExpire).Unix())
		conn.Do("EXEC")
	}

	{ // live view count. Rolling 30s window expiry
		k := s.productDataKey(productID, "live")
		n := time.Now().Unix()
		d := n - int64(ProductLiveExpire.Seconds())
		conn.Send("MULTI")
		conn.Send("ZADD", k, n, userID)
		conn.Send("ZREMRANGEBYSCORE", k, 0, d)
		conn.Do("EXEC")
	}

	return nil
}

func (s *Stash) IncrProductPurchases(productID int64) (int, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int(conn.Do("INCR", s.productDataKey(productID, "buys")))
}

func (s *Stash) productDataKey(productID int64, parts ...string) string {
	return s.buildKey("product", strconv.FormatInt(productID, 10), parts...)
}
