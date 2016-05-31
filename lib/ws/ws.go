package ws

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

func Bind(payload io.Reader, v interface{}) error {
	err := json.NewDecoder(payload).Decode(v)
	io.Copy(ioutil.Discard, payload)
	if err != nil {
		return err
	}
	return nil
}

func Respond(w http.ResponseWriter, status int, v interface{}) {
	if err, ok := v.(error); ok {
		JSON(w, status, WrapError(err))
		return
	}
	val := reflect.ValueOf(v)

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
