package builder

import (
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/schmurfy/chipi/schema"
	"github.com/schmurfy/chipi/wrapper"
)

type rawHandler interface {
	Handle(http.ResponseWriter, *http.Request)
}

type Builder struct {
	swagger *openapi3.T
	schema  *schema.Schema
}

func New(infos *openapi3.Info) (*Builder, error) {
	swagger := &openapi3.T{
		OpenAPI: "3.1.0",
		Info:    infos,
	}

	s, err := schema.New()
	if err != nil {
		return nil, err
	}

	ret := &Builder{
		swagger: swagger,
		schema:  s,
	}

	return ret, nil
}

func (b *Builder) AddTag(tag *openapi3.Tag) {
	b.swagger.Tags = append(b.swagger.Tags, tag)
}

func (b *Builder) AddServer(server *openapi3.Server) {
	b.swagger.AddServer(server)
}

func (b *Builder) AddSecurityScheme(name string, s *openapi3.SecurityScheme) {
	if b.swagger.Components.SecuritySchemes == nil {
		b.swagger.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
	}

	b.swagger.Components.SecuritySchemes[name] = &openapi3.SecuritySchemeRef{
		Value: s,
	}
}

func (b *Builder) AddSecurityRequirement(req openapi3.SecurityRequirement) {
	b.swagger.Security.With(req)
}

func (b *Builder) ServeSchema(w http.ResponseWriter, r *http.Request) {
	data, err := b.swagger.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

type CallbackFunc func(http.ResponseWriter, interface{})

func (b *Builder) Get(r chi.Router, pattern string, reqObject interface{}) (err error) {
	_, err = b.Method(r, pattern, "GET", reqObject)
	return
}

func (b *Builder) Post(r chi.Router, pattern string, reqObject interface{}) (err error) {
	_, err = b.Method(r, pattern, "POST", reqObject)
	return
}

func (b *Builder) Patch(r chi.Router, pattern string, reqObject interface{}) (err error) {
	_, err = b.Method(r, pattern, "PATCH", reqObject)
	return
}

func (b *Builder) Delete(r chi.Router, pattern string, reqObject interface{}) (err error) {
	_, err = b.Method(r, pattern, "DELETE", reqObject)
	return
}

func (b *Builder) Method(r chi.Router, pattern string, method string, reqObject interface{}) (op *openapi3.Operation, err error) {
	op = openapi3.NewOperation()

	// analyze parameters if any
	typ := reflect.TypeOf(reqObject)
	if (typ.Kind() != reflect.Ptr) || (typ.Elem().Kind() != reflect.Struct) {
		err = errors.New("wrong type, pointer to struct expected")
		return
	}

	typ = typ.Elem()
	op.OperationID = typ.Name()

	if _, ok := reqObject.(wrapper.HandlerInterface); ok {
		r.Method(method, pattern, wrapper.WrapRequest(reqObject))
	} else if rr, ok := reqObject.(rawHandler); ok {
		r.Method(method, pattern, http.HandlerFunc(rr.Handle))
	} else {
		err = errors.Errorf("%T object must implement HandlerInterface interface", reqObject)
		return
	}

	err = b.generateOperationDoc(r, op, typ)
	if err != nil {
		return
	}

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

	// Headers
	err = b.generateHeadersDoc(r, op, typ)
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
