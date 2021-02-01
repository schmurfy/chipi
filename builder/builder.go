package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/go-chi/chi"
)

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

func (b *Builder) Get(r chi.Router, pattern string, reqObject interface{}, h CallbackFunc) error {
	return b.method(r, pattern, "GET", reqObject, h)
}

func (b *Builder) Post(r chi.Router, pattern string, reqObject interface{}, h CallbackFunc) error {
	return b.method(r, pattern, "POST", reqObject, h)
}

func (b *Builder) method(r chi.Router, pattern string, method string, reqObject interface{}, h CallbackFunc) error {
	op := openapi3.NewOperation()
	// op.Description = ""...""

	r.Method(method, pattern, wrapRequest(reqObject, h))

	// analyze parameters if any
	typ := reflect.TypeOf(reqObject)
	if typ.Kind() != reflect.Struct {
		return errors.New("wrong type, struct expected")
	}

	// URL Parameters
	err := b.generateParametersDoc(r, op, typ, method)
	if err != nil {
		return err
	}

	// body
	err = b.generateBodyDocumentation(op, typ)
	if err != nil {
		return err
	}

	// response
	err = b.generateResponseDocumentation(op, typ)
	if err != nil {
		return err
	}

	b.swagger.AddOperation(pattern, method, op)
	return nil
}

func wrapRequest(req interface{}, h CallbackFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: add checks
		vv := reflect.New(reflect.TypeOf(req)).Elem()

		// path
		pathValue := vv.FieldByName("Path")

		rctx := chi.RouteContext(r.Context())
		for _, k := range rctx.URLParams.Keys {
			fieldValue := pathValue.FieldByName(k)
			if fieldValue.IsValid() {
				fieldValue.SetString(rctx.URLParam(k))
			}
		}

		// body
		// bodyValue := vv.FieldByName("Body")
		// if !bodyValue.IsZero() {
		// 	decoder := json.NewDecoder(r.Body)
		// 	obj := bodyValue.Interface()
		// 	err := decoder.Decode(&obj)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }

		h(w, vv.Addr().Interface())
	}
}

func (b *Builder) generateParametersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type, method string) error {
	pathField, found := requestObjectType.FieldByName("Path")
	if !found {
		return errors.New("wrong struct, Path field expected")
	}

	example := pathField.Tag.Get("example")

	tctx := chi.NewRouteContext()
	if r.Match(tctx, method, example) {
		gen := openapi3gen.NewGenerator()

		for _, key := range tctx.URLParams.Keys {
			// pathStruct must contain all defined keys
			paramField, found := pathField.Type.FieldByName(key)
			if !found {
				return fmt.Errorf("wrong path struct, field %s expected", key)
			}

			schema, err := gen.GenerateSchemaRef(paramField.Type)
			if err != nil {
				return err
			}

			param := openapi3.NewPathParameter(key).
				WithSchema(schema.Value)

			paramsExample, found := paramField.Tag.Lookup("example")
			if found {
				param.Example = paramsExample
			}

			op.AddParameter(param)
		}
		fmt.Printf("MATCH: %v\n", tctx.URLParams)
	}

	return nil
}

func schemaFromType(typ reflect.Type) (*openapi3.Schema, error) {
	gen := openapi3gen.NewGenerator()
	schema, err := gen.GenerateSchemaRef(typ)
	if err != nil {
		return nil, err
	}

	for ref := range gen.SchemaRefs {
		ref.Ref = ""
	}

	return schema.Value, nil
}

func (b *Builder) generateBodyDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	bodyField, found := requestObjectType.FieldByName("Body")
	if found {
		body := openapi3.NewRequestBody()

		bodySchema, err := schemaFromType(bodyField.Type)
		if err != nil {
			return err
		}

		data, err := json.Marshal(bodySchema)
		if err != nil {
			panic(err)
		}
		fmt.Printf("schema %v : %s\n", bodyField.Type, data)

		bodyRef := &openapi3.RequestBodyRef{
			// Ref:   "#/components/requestBodies/pet",
			Value: body,
		}

		body.Content = openapi3.NewContentWithJSONSchema(bodySchema)
		op.RequestBody = bodyRef
	}

	return nil
}

func (b *Builder) generateResponseDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		responseSchema, err := schemaFromType(responseField.Type)
		if err != nil {
			return err
		}

		resp.Content = openapi3.NewContentWithJSONSchema(responseSchema)

		op.Responses = make(openapi3.Responses)
		op.Responses["200"] = &openapi3.ResponseRef{
			Value: resp,
		}
	}

	return nil
}
