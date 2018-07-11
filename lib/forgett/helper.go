package forgett

import (
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// Float64Map converts a redis result of key-value pairs into a map of string-float64 pairs.
func Float64Map(result interface{}, err error) (map[string]float64, error) {
	values, err := redis.Values(result, err)
	if err != nil {
		return nil, err
	}

	if len(values)%2 != 0 {
		return nil, errors.New("float64Map() expects even number of values result")
	}

	m := make(map[string]float64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].([]byte)
		if !ok {
			return nil, errors.New("key not a bulk string value")
		}
		value, err := redis.Float64(values[i+1], nil)
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}
