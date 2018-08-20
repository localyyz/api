package forgett

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type Distribution struct {
	Name               string
	primary, secondary *Set
}

func NewDistribution(name string, lifetime time.Duration, norm int64) (*Distribution, error) {
	d, err := FetchDistribution(name, lifetime, norm)
	if err != nil {
		return nil, err
	}

	if exists, _ := d.Exists(); !exists {
		d.primary.Init()
		d.secondary.Init()
	}

	return d, nil
}

func FetchDistribution(name string, lifetime time.Duration, norm int64) (*Distribution, error) {
	d := &Distribution{
		Name: name,
	}
	// create a primary and secondary set
	d.primary, _ = NewSet(d.Name, lifetime, 1)
	d.secondary, _ = NewSet(
		fmt.Sprintf("%s_%dt", d.Name, norm),
		lifetime,
		norm,
	)
	return d, nil
}

func (d *Distribution) Exists() (bool, error) {
	conn := DefaultClient.pool.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("EXISTS", d.primary.Name))
}

func (d *Distribution) TopScores(limit, offset int) (map[string]float64, error) {
	counts, err := d.primary.Scores(limit, offset)
	if err != nil {
		return nil, err
	}
	norm, err := d.secondary.Scores(limit, offset)
	if err != nil {
		return nil, err
	}
	return normalizeScores(counts, norm), nil
}

// Scores returns the trending scores for all entries
func (d *Distribution) Scores() (map[string]float64, error) {
	counts, err := d.primary.AllScores()
	if err != nil {
		return nil, err
	}
	norm, err := d.secondary.AllScores()
	if err != nil {
		return nil, err
	}
	return normalizeScores(counts, norm), nil
}

func normalizeScores(counts, norm map[string]float64) map[string]float64 {
	// normalize member scores
	result := make(map[string]float64, len(counts))
	for k, v := range counts {
		normV := float64(norm[k])
		if normV == 0.0 {
			result[k] = 0.0
		} else {
			result[k] = float64(v) / float64(normV)
		}
	}
	return result
}

func (d *Distribution) Incr(bin string) error {
	return d.IncrBy(bin, 1)
}

func (d *Distribution) IncrBy(bin string, by int64) error {
	if err := d.primary.IncrBy(bin, by); err != nil {
		return errors.Wrap(err, "increment primary")
	}
	if err := d.secondary.IncrBy(bin, by); err != nil {
		return errors.Wrap(err, "increment secondary")
	}
	return nil
}
