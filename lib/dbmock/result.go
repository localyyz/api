package dbmock

import (
	"reflect"

	db "upper.io/db.v3"
)

type mockResult struct {
	// slice of fixture values
	//	-> map[reflect.Type][]reflect.Value
	f fixtures
}

var _ interface {
	db.Result
} = &mockResult{}

// db.Result
func (m mockResult) String() string { return "" }

func (m mockResult) Limit(int) db.Result { return m }

func (m mockResult) Offset(int) db.Result { return m }

func (m mockResult) OrderBy(...interface{}) db.Result { return m }

func (m mockResult) Select(...interface{}) db.Result { return m }

func (m mockResult) Where(...interface{}) db.Result { return m }

func (m mockResult) And(...interface{}) db.Result { return m }

func (m mockResult) Group(...interface{}) db.Result { return m }

func (m mockResult) Delete() error { return nil }

func (m mockResult) Update(interface{}) error { return nil }

func (m mockResult) Count() (uint64, error) { return 0, nil }

func (m mockResult) Exists() (bool, error) { return false, nil }

func (m mockResult) Next(ptrToStruct interface{}) bool { return false }
func (m mockResult) Err() error                        { return nil }

func (m mockResult) One(ptrToStruct interface{}) error {
	dstv := reflect.ValueOf(ptrToStruct)
	itemV := dstv.Elem()

	itemT := itemV.Type()
	objT := itemT.Elem()

	var item reflect.Value
	if m.f != nil && len(m.f) != 0 && m.f[itemT] != nil {
		for _, v := range m.f[itemT] {
			item = v
			break
		}
	} else {
		item = reflect.New(objT)
	}
	itemV.Set(item)

	if itemV.IsNil() {
		return db.ErrNoMoreRows
	}
	return nil
}

func (m mockResult) All(sliceOfStructs interface{}) error {
	dstv := reflect.ValueOf(sliceOfStructs)
	slicev := dstv.Elem()
	itemT := slicev.Type().Elem()

	var item reflect.Value
	if m.f != nil && len(m.f) > 0 && m.f[itemT] != nil {
		slicev = reflect.Append(slicev, m.f[itemT]...)
	} else {
		item = reflect.New(itemT.Elem())
		slicev = reflect.Append(slicev, item)
	}

	dstv.Elem().Set(slicev)

	return nil
}

func (m mockResult) Paginate(pageSize uint) db.Result { return m }

func (m mockResult) Page(pageNumber uint) db.Result { return m }

func (m mockResult) Cursor(cursorColumn string) db.Result { return m }

func (m mockResult) NextPage(cursorValue interface{}) db.Result { return m }

func (m mockResult) PrevPage(cursorValue interface{}) db.Result { return m }

func (m mockResult) TotalPages() (uint, error) { return 0, nil }

func (m mockResult) TotalEntries() (uint64, error) { return 0, nil }

func (m mockResult) Close() error { return nil }
