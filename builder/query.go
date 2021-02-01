package builder

import (
	"errors"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateQueryParametersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type) error {
	pathField, found := requestObjectType.FieldByName("Query")
	if !found {
		return errors.New("wrong struct, Query field expected")
	}

	queryStructType := pathField.Type
	if queryStructType.Kind() != reflect.Struct {
		return errors.New("expected struct for Query")
	}

	for i := 0; i < queryStructType.NumField(); i++ {
		field := queryStructType.Field(i)
		gen := schema.NewGenerator()

		schema, err := gen.GenerateSchemaRef(field.Type)
		if err != nil {
			return err
		}

		param := openapi3.NewQueryParameter(field.Name).
			WithSchema(schema.Value)

		paramsExample, found := field.Tag.Lookup("example")
		if found {
			param.Example = paramsExample
		}

		op.AddParameter(param)
	}

	return nil
}
