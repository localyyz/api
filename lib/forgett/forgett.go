package forgett

// heavily inspired https://github.com/bitly/forgettable

import (
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type Forgett struct {
	// TODO make configurable
	opts Options
	pool *redis.Pool
}

type Options struct {
	// normalizing modifier
	Norm int64

	// Default lifetime
	Lifetime time.Duration
}

var (
	ErrRedisPoolInit     = errors.New("forgett: redis pool not initiated")
	ErrDistributionEmpty = errors.New("forgett: skip update distribution empty")
	ErrDistributionField = errors.New("forgett: cannot fetch field")

	Day = 24 * time.Hour

	DefaultClient *Forgett
)

var DefaultOptions = Options{
	Norm:     2,
	Lifetime: 30 * time.Second,
}

type Option func(*Options) error

func NormMod(v int64) Option {
	return func(o *Options) error {
		o.Norm = v
		return nil
	}
}

func NewForgett(pool *redis.Pool, options ...Option) (*Forgett, error) {
	fgt := &Forgett{
		opts: DefaultOptions,
		pool: pool,
	}
	for _, opt := range options {
		if err := opt(&fgt.opts); err != nil {
			return nil, err
		}
	}
	DefaultClient = fgt
	return fgt, nil
}

func init() {
	// seed
	rand.Seed(time.Now().UnixNano())
}
