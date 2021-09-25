package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func (b *Builder) generateQueryParametersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type) error {
	pathField, found := requestObjectType.FieldByName("Query")
	if !found {
		return nil
	}

	queryStructType := pathField.Type
	if queryStructType.Kind() != reflect.Struct {
		return errors.New("expected struct for Query")
	}

	for i := 0; i < queryStructType.NumField(); i++ {
		field := queryStructType.Field(i)

		schema, err := b.schema.GenerateSchemaFor(b.swagger, field.Type)
		if err != nil {
			return err
		}

		param := openapi3.NewQueryParameter(field.Name).
			WithSchema(schema.Value)

		err = fillParamFromTags(requestObjectType, param, field, "Query")
		if err != nil {
			return err
		}

		op.AddParameter(param)
	}

	return nil
}
