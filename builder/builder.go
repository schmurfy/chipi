package builder

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"

	"github.com/schmurfy/chipi/schema"
)

type RequestInterface interface {
	Handle(http.ResponseWriter)
}

type Builder struct {
	swagger *openapi3.Swagger
}

func New(infos *openapi3.Info) (*Builder, error) {
	swagger := &openapi3.Swagger{
		OpenAPI: "3.0.0",
		Info:    infos,
	}

	ret := &Builder{
		swagger: swagger,
	}

	return ret, nil
}

func (b *Builder) AddServer(server *openapi3.Server) {
	b.swagger.AddServer(server)
}

func (b *Builder) ServeSchema(w http.ResponseWriter, r *http.Request) {
	data, err := b.swagger.MarshalJSON()
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

type CallbackFunc func(http.ResponseWriter, interface{})

func (b *Builder) Get(r chi.Router, pattern string, reqObject RequestInterface) (op *openapi3.Operation, err error) {
	return b.method(r, pattern, "GET", reqObject)
}

func (b *Builder) Post(r chi.Router, pattern string, reqObject RequestInterface) (op *openapi3.Operation, err error) {
	return b.method(r, pattern, "POST", reqObject)
}

func (b *Builder) method(r chi.Router, pattern string, method string, reqObject RequestInterface) (op *openapi3.Operation, err error) {
	op = openapi3.NewOperation()
	// op.Description = ""...""

	// analyze parameters if any
	typ := reflect.TypeOf(reqObject)
	fmt.Printf("type: %v\n", typ.Kind())
	if (typ.Kind() != reflect.Ptr) || (typ.Elem().Kind() != reflect.Struct) {
		err = errors.New("wrong type, pointer to struct expected")
		return
	}

	typ = typ.Elem()
	r.Method(method, pattern, wrapRequest(typ))

	// URL Parameters
	err = b.generateParametersDoc(r, op, typ, method)
	if err != nil {
		return
	}

	// Query parameters
	err = b.generateQueryParametersDoc(r, op, typ)
	if err != nil {
		return
	}

	// body
	err = b.generateBodyDocumentation(op, typ)
	if err != nil {
		return
	}

	// response
	err = b.generateResponseDocumentation(op, typ)
	if err != nil {
		return
	}

	b.swagger.AddOperation(pattern, method, op)
	return
}

func setFieldValue(f reflect.Value, value string) error {
	switch f.Type().Kind() {
	case reflect.Int:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		f.SetInt(n)
	}

	return nil
}

func schemaFromType(typ reflect.Type) (*openapi3.Schema, error) {
	g := schema.NewGenerator()
	schema, err := g.GenerateSchemaRef(typ)
	if err != nil {
		return nil, err
	}

	for ref := range g.SchemaRefs {
		ref.Ref = ""
	}

	return schema.Value, nil
}
