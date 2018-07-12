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

	return redis.Int64(conn.Do("ZSCORE", "product:views", productID))
}

func (s *Stash) GetProductLiveViews(productID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("ZCARD", s.productDataKey(productID, "live")))
}

func (s *Stash) GetProductPurchases(productID int64) (int64, error) {
	conn := s.conn()
	defer conn.Close()

	return redis.Int64(conn.Do("ZSCORE", "product:buys", productID))
}

func (s *Stash) IncrProductViews(productID, userID int64) error {
	conn := s.conn()
	defer conn.Close()

	// count the view
	conn.Do("ZADD", "product:views", 1, productID)

	{ // live view count. Rolling 15min window expiry
		k := s.productDataKey(productID, "live")
		n := time.Now().Unix()
		d := n - int64(ProductLiveExpire.Seconds())
		conn.Send("MULTI")
		conn.Send("ZADD", k, n, userID)
		// TODO: this should be a cronjob cleanup
		conn.Send("ZREMRANGEBYSCORE", k, 0, d)
		conn.Do("EXEC")
	}

	return nil
}

func (s *Stash) IncrProductPurchases(productID int64) (int, error) {
	conn := s.conn()
	defer conn.Close()

	// count the view
	return redis.Int(conn.Do("ZADD", "product:buys", 1, productID))
}

func (s *Stash) productDataKey(productID int64, parts ...string) string {
	return s.buildKey("product", strconv.FormatInt(productID, 10), parts...)
}
