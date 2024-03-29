package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

// Result is the intermediate data type to decode
// the incoming request body into
type Result map[string]interface{}

// requiredDecoder checks v struct tags and
// validates any and all `required` fields are satisfied
// by the incoming payload
func requiredDecoder(r *http.Request, v interface{}) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}

			err = r.(error)
		}
		if err != nil {
			lg.Error(err)
		}
	}()

	// decode the json into a placeholder result map
	var b []byte
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var rl Result
	err = json.Unmarshal(b, &rl)
	if err != nil {
		return err
	}

	// check required fields
	err = checkRequired(rl, reflect.ValueOf(v))
	if err != nil {
		return err
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
		links = append(links, fmt.Sprintf("<%s>; rel=\"%s\"", url, name))
	}
	w.Header().Set("Link", strings.Join(links, ","))
}

func renderResponder(w http.ResponseWriter, r *http.Request, v interface{}) {
	if err, ok := v.(error); ok {
		werr := WrapErr(err)
		w.WriteHeader(werr.StatusCode)
		render.DefaultResponder(w, r, werr)
		return
	}
	// pagination Link header
	if cursor, _ := r.Context().Value("cursor").(*Page); cursor != nil {
		cursorLinkHeader(w, cursor)
		w.Header().Set("X-Item-Total", fmt.Sprintf("%d", cursor.ItemTotal))
	}
	render.DefaultResponder(w, r, v)
}

func init() {
	// inject and override defaults in the render package
	render.Decode = requiredDecoder
	render.Respond = renderResponder
}
