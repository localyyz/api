package dbmock

import (
	"reflect"
)

const FixtureCtxKey = "dbmock.fixtures"

// TODO: reset the fixture
type fixtures map[reflect.Type][]reflect.Value

func NewFixture() fixtures {
	return fixtures{}
}

func (f fixtures) Add(ptrToStruct interface{}) {
	dstv := reflect.ValueOf(ptrToStruct) // pointer value
	itemT := dstv.Type()
	f[itemT] = append(f[itemT], dstv)
}
