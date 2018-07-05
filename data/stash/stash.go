// Stash
package stash

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

const ResultLimit = 50

var (
	DefaultClient           *Stash
	ErrDefaultNotConfigured = errors.New("stash: DefaultClient has not been configured")
)

type Stash struct {
	pool *redis.Pool
}

func NewDialer(host string) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		return redis.Dial("tcp", host)
	}
}

func NewPool(dial func() (redis.Conn, error)) *redis.Pool {
	//TODO: make these configurable
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   50,
		IdleTimeout: 300 * time.Second,
		Wait:        true,
		Dial:        dial,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func NewStash(host string) (*Stash, error) {
	pool := NewPool(NewDialer(host))
	return &Stash{pool: pool}, nil
}

func SetupDefaultClient(host string) (err error) {
	DefaultClient, err = NewStash(host)
	return err
}

func Ping() error {
	if DefaultClient == nil {
		return ErrDefaultNotConfigured
	}
	return DefaultClient.Ping()
}

func (s *Stash) Ping() error {
	conn := s.conn()
	defer conn.Close()

	reply, err := conn.Do("PING")
	if err != nil {
		return err
	}

	val, ok := reply.(string)
	if !ok || val != "PONG" {
		return errors.New("stash: service unreachable")
	}
	return nil
}

func (s *Stash) conn() redis.Conn {
	return s.pool.Get()
}

func (s *Stash) buildKey(prefix string, id string, parts ...string) string {
	key := fmt.Sprintf("%s:%s", prefix, id)
	if len(parts) > 0 {
		return key + ":" + strings.Join(parts, ":")
	} else {
		return key
	}
}

func (s *Stash) bucketKey(bucketId string, parts ...string) string {
	return s.buildKey("bucket", bucketId, parts...)
}
