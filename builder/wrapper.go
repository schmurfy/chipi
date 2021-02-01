package builder

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/go-chi/chi"
)

func wrapRequest(typ reflect.Type) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: add checks
		vv := reflect.New(typ).Elem()

		// path
		pathValue := vv.FieldByName("Path")

		rctx := chi.RouteContext(r.Context())
		for _, k := range rctx.URLParams.Keys {
			fieldValue := pathValue.FieldByName(k)
			if fieldValue.IsValid() {
				fieldValue.SetString(rctx.URLParam(k))
			}
		}

		// query
		queryValue := vv.FieldByName("Query")
		for k, v := range r.URL.Query() {
			f := queryValue.FieldByName(k)
			if f.IsValid() {
				err := setFieldValue(f, v[0])
				if err != nil {
					panic(err)
				}
			}
		}

		// body
		bodyValue := vv.FieldByName("Body")
		if bodyValue.IsValid() {
			// body, err := ioutil.ReadAll(r.Body)
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Printf("body: %s\n", body)
			decoder := json.NewDecoder(r.Body)
			obj := bodyValue.Addr().Interface()
			err := decoder.Decode(&obj)
			if err != nil {
				panic(err)
			}
		}

		var filledRequestObject RequestInterface = vv.Addr().Interface().(RequestInterface)
		filledRequestObject.Handle(w)
	}
}
