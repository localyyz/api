package forgett

import (
	"math"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

const (
	LastDecayKey         = "_lastdecay"
	scrubFilter  float64 = 0.0001
)

// Set is a collection of observables stored in redis
// with an average observation lifetime and a start time.
type Set struct {
	Name string
	// modifier is the rate to multiply lifetime by
	modifier int64

	lifetime time.Duration

	lastDecayKey string
}

func NewSet(name string, lifetime time.Duration, mod int64) (*Set, error) {
	s := &Set{
		Name:         name,
		lifetime:     lifetime,
		modifier:     mod,
		lastDecayKey: name + LastDecayKey,
	}
	return s, nil
}

func (s *Set) Init() error {
	// TODO: check errors
	s.UpdateDecayTime()

	return nil
}

func (s *Set) AllScores() (map[string]float64, error) {
	return s.Scores(-1)
}

func (s *Set) Scores(limit int) (map[string]float64, error) {
	if err := s.decay(); err != nil {
		return nil, err
	}
	if err := s.scrub(); err != nil {
		return nil, err
	}
	return s.FetchN(limit)
}

func (s *Set) decay() error {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	// get our parameters for decay
	lastDecayedAt, err := s.LastDecayTime()
	if err != nil {
		return err
	}

	deltaTime := time.Now().UTC().Sub(lastDecayedAt).Seconds()

	// loop through our entries
	conn.Send("MULTI")
	m, _ := s.Fetch()
	for k, v := range m {
		decay := float64(v) * math.Exp(-deltaTime/s.lifetime.Seconds())
		conn.Send("ZADD", s.Name, decay, k)
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return errors.Wrap(err, "forgett: decay")
	}

	return s.UpdateDecayTime()
}

func (s *Set) scrub() error {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYSCORE", s.Name, "-inf", scrubFilter)
	return err
}

// Fetch retrieves scores from highest to lowest
func (s *Set) Fetch() (map[string]float64, error) {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	return s.FetchN(-1)
}

// Fetch retrieves the first <limit> number of scores from highest to lowest
func (s *Set) FetchN(limit int) (map[string]float64, error) {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	return Float64Map(conn.Do("ZREVRANGE", s.Name, 0, limit, "WITHSCORES"))
}

// Incr increments the member value by 1
func (s *Set) Incr(bin string) {
	s.IncrBy(bin, 1)
}

// Incr increments the member value by the given value
func (s *Set) IncrBy(bin string, by int64) error {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	_, err := conn.Do("ZINCRBY", s.Name, by, bin)
	return err
}

// LastDecayDate returns the datetime of the last decay
func (s *Set) LastDecayTime() (time.Time, error) {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	seconds, err := redis.Int64(conn.Do("GET", s.lastDecayKey))
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}

func (s *Set) UpdateDecayTime() error {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", s.lastDecayKey, time.Now().Unix())
	return err
}
