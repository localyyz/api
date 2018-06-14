package dbmock

import (
	"context"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type mockTable struct {
	name   string
	result mockResult
	calls  []call
}

type call string

// check if a function have been called
func (m mockTable) HasCalledMethod(name string, n int) bool {
	var count int
	for _, n := range m.calls {
		if n == call(name) {
			count++
		}
	}
	return count == n
}

// bond.Store
func (m mockTable) Session() bond.Session                    { return nil }
func (m mockTable) WithSession(sess bond.Session) bond.Store { return nil }

func (m mockTable) Save(interface{}) error {
	m.calls = append(m.calls, call("Save"))
	return nil
}
func (m mockTable) Delete(interface{}) error {
	m.calls = append(m.calls, call("Delete"))
	return nil
}
func (m mockTable) Update(interface{}) error {
	m.calls = append(m.calls, call("Update"))
	return nil
}
func (m mockTable) Create(interface{}) error {
	m.calls = append(m.calls, call("Create"))
	return nil
}

// db.Collection
func (m mockTable) Insert(interface{}) (interface{}, error) { return nil, nil }
func (m mockTable) InsertReturning(interface{}) error       { return nil }
func (m mockTable) UpdateReturning(interface{}) error       { return nil }
func (m mockTable) Exists() bool                            { return false }
func (m mockTable) Find(...interface{}) db.Result {
	return m.result
}
func (m mockTable) Truncate() error { return nil }
func (m mockTable) Name() string    { return m.name }

// table store
type mockTableStore struct {
	bond.Store
}

func NewTable(ctx context.Context) *mockTable {
	var m mockResult
	if f, ok := ctx.Value(FixtureCtxKey).(fixtures); ok {
		m = mockResult{f}
	}
	return &mockTable{result: m}
}
