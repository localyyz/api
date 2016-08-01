package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Result map[string]interface{}

// Bind is a custom json decoder that adds additional json tag such as
// `required`
func Bind(payload io.Reader, v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}

			err = r.(error)
		}
	}()

	// decode the json into a placeholder result map
	b, err := ioutil.ReadAll(payload)
	if err != nil {
		return
	}

	var r Result
	err = json.Unmarshal(b, &r)
	if err != nil {
		return
	}

	// check required fields
	err = checkRequired(r, reflect.ValueOf(v))
	if err != nil {
		return
	}

	// finally, decode into v
	return json.Unmarshal(b, &v)
}

func checkRequired(r Result, v reflect.Value) error {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("output value must be a struct.")
	}

	if !v.CanSet() {
		return fmt.Errorf("output value cannot be set.")
	}

	vType := v.Type()
	num := vType.NumField()

	for i := 0; i < num; i++ {

		jsonTag := vType.Field(i).Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		index := strings.IndexRune(jsonTag, ',')
		var name string
		if index == -1 {
			name = jsonTag
		} else {
			name = jsonTag[:index]
			if jsonTag[index:] == ",required" {
				// if required field and not found, throw error
				if _, ok := r[name]; !ok {
					return fmt.Errorf("required field '%v' missing in request", name)
				}
			}
		}
	}

	return nil
}

func cursorLinkHeader(w http.ResponseWriter, cursor *Page) {
	var links []string
	for name, url := range cursor.PageURLs() {
		links = append(links, fmt.Sprintf("<%s>;rel=\"%s\"", url, name))
	}

	w.Header().Set("Link", strings.Join(links, ","))
}

func Respond(w http.ResponseWriter, status int, v interface{}, optCursor ...*Page) {
	if err, ok := v.(error); ok {
		status, err = WrapError(status, err)
		JSON(w, status, err.Error())
		return
	}
	val := reflect.ValueOf(v)

	if len(optCursor) > 0 && optCursor[0] != nil {
		cursor := optCursor[0]
		cursorLinkHeader(w, cursor)

		// Return first element of the slice only.
		if cursor.FirstOnly() {
			if val.Kind() == reflect.Slice {
				if val.Len() > 0 {
					v = val.Index(0).Interface()
				} else {
					v = nil
				}
			}
			JSON(w, status, v)
			return
		}
	}

	// Force to return empty JSON array [] instead of null in case of zero slice.
	if val.Kind() == reflect.Slice && val.IsNil() {
		v = reflect.MakeSlice(val.Type(), 0, 0).Interface()
	}

	JSON(w, status, v)
}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(b) > 0 {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(b)
}
