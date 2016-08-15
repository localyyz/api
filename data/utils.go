package data

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"upper.io/db.v2"
)

var (
	ErrPromoStart = errors.New("promo have not started")
	ErrPromoEnded = errors.New("promo ended")
	ErrPromoUsed  = errors.New("promo already used")
	ErrPromoPlace = errors.New("promo cannot be applied to this place")
)

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
