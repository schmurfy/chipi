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

type Method struct {
	pattern   string
	method    string
	reqObject interface{}
}
type Builder struct {
	swagger *openapi3.T
	schema  *schema.Schema
	router  *chi.Mux
	methods []*Method
}

func New(r *chi.Mux, infos *openapi3.Info) (*Builder, error) {
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
		router:  r,
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
	data, err := b.GenerateJson()
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
	err = b.Method(r, pattern, "GET", reqObject)
	return
}

func (b *Builder) Post(r chi.Router, pattern string, reqObject interface{}) (err error) {
	err = b.Method(r, pattern, "POST", reqObject)
	return
}

func (b *Builder) Patch(r chi.Router, pattern string, reqObject interface{}) (err error) {
	err = b.Method(r, pattern, "PATCH", reqObject)
	return
}

func (b *Builder) Delete(r chi.Router, pattern string, reqObject interface{}) (err error) {
	err = b.Method(r, pattern, "DELETE", reqObject)
	return
}

func (b *Builder) findRoute(typ reflect.Type, method string) (*chi.Context, error) {
	pathField, found := typ.FieldByName("Path")
	if !found {
		return nil, errors.New("Path field not found")
	}

	routeExample, found := pathField.Tag.Lookup("example")
	if !found {
		return nil, errors.New("example tag not found")
	}

	tctx := chi.NewRouteContext()
	if b.router.Match(tctx, method, routeExample) {
		return tctx, nil
	}

	return nil, errors.New("route not found")
}

func (b *Builder) Method(r chi.Router, pattern string, method string, reqObject interface{}) (err error) {

	if _, ok := reqObject.(wrapper.HandlerInterface); ok {
		r.Method(method, pattern, wrapper.WrapRequest(reqObject))
	} else if rr, ok := reqObject.(rawHandler); ok {
		r.Method(method, pattern, http.HandlerFunc(rr.Handle))
	} else {
		err = errors.Errorf("%T object must implement HandlerInterface interface", reqObject)
		return
	}

	b.methods = append(b.methods, &Method{
		pattern:   pattern,
		method:    method,
		reqObject: reqObject,
	})
	return
}

func (b *Builder) GenerateJson() ([]byte, error) {

	swagger := *b.swagger
	for _, m := range b.methods {
		op := openapi3.NewOperation()

		// analyze parameters if any
		typ := reflect.TypeOf(m.reqObject)
		if (typ.Kind() != reflect.Ptr) || (typ.Elem().Kind() != reflect.Struct) {
			err := errors.New("wrong type, pointer to struct expected")
			return nil, err
		}

		typ = typ.Elem()
		op.OperationID = typ.Name()

		routeContext, err := b.findRoute(typ, m.method)
		if routeContext == nil {
			return nil, err
		}

		err = generateOperationDoc(op, typ)
		if err != nil {
			return nil, err
		}

		// URL Parameters
		err = b.generateParametersDoc(&swagger, op, typ, m.method, routeContext)
		if err != nil {
			return nil, err
		}

		// Query parameters
		err = b.generateQueryParametersDoc(&swagger, op, typ)
		if err != nil {
			return nil, err
		}

		// Headers
		err = b.generateHeadersDoc(&swagger, op, typ)
		if err != nil {
			return nil, err
		}

		// body
		err = b.generateBodyDoc(&swagger, op, m.reqObject, typ)
		if err != nil {
			return nil, err
		}

		// response
		err = b.generateResponseDoc(&swagger, op, m.reqObject, typ)
		if err != nil {
			return nil, err
		}

		swagger.AddOperation(routeContext.RoutePattern(), m.method, op)

	}

	json, err := swagger.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json, nil
}
