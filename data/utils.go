package data

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/goware/geotools"

	"upper.io/db.v3"
)

const earthRadiusInMeters = 6378100
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// DistanceTo returns distance between two geo locations using the Haversine formula
// Reference: https://gist.github.com/cdipaolo/d3f8db3848278b49db68
func DistanceTo(start, dst *geotools.LatLng) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2 float64
	la1 = start.Lat * math.Pi / 180
	lo1 = start.Lng * math.Pi / 180
	la2 = dst.Lat * math.Pi / 180
	lo2 = dst.Lng * math.Pi / 180

	// calculate
	dla := math.Sin(0.5 * (la2 - la1))
	dlo := math.Sin(0.5 * (lo2 - lo1))
	h := dla*dla + math.Cos(la1)*math.Cos(la2)*dlo*dlo

	return 2 * earthRadiusInMeters * math.Asin(math.Sqrt(h))
}

func GetTimeUTCPointer() *time.Time {
	t := time.Now().UTC()
	return &t
}

func WordLimit(s string, words int) string {
	parts := strings.Fields(s)
	if len(parts) <= words {
		// Join again, because this can possibly create a higher quality string - Fields remove double whitespace.
		return strings.Join(parts, " ")
	}
	return strings.Join(parts[0:words], " ") + " ..."
}

func MaintainOrder(field string, customOrdering interface{}) db.RawValue {
	if reflect.TypeOf(customOrdering).Kind() != reflect.Slice {
		panic("customOrdering is not a slice")
	}

	vals := reflect.ValueOf(customOrdering)
	if vals.Len() == 0 {
		return db.Raw("true")
	}

	sort := make([]string, vals.Len()+2)
	sort[0] = "CASE"
	sort[len(sort)-1] = "END"

	switch reflect.TypeOf(customOrdering).Elem().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		for i := 0; i < vals.Len(); i++ {
			sort[i+1] = fmt.Sprintf("WHEN \"%s\" = %v THEN %d", field, vals.Index(i), i)
		}

	default:
		panic("customOrdering is not a slice of int (any) or float (any)")
	}

	return db.Raw(strings.Join(sort, " "))
}

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}
