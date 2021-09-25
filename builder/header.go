package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func (b *Builder) generateHeadersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type) error {
	headerField, found := requestObjectType.FieldByName("Header")
	if !found {
		return nil
	}

	headerStructType := headerField.Type
	if headerStructType.Kind() != reflect.Struct {
		return errors.New("expected struct for Header")
	}

	for i := 0; i < headerStructType.NumField(); i++ {
		field := headerStructType.Field(i)

		schema, err := b.schema.GenerateSchemaFor(b.swagger, field.Type)
		if err != nil {
			return err
		}

		param := openapi3.NewHeaderParameter(field.Name).
			WithSchema(schema.Value)

		err = fillParamFromTags(requestObjectType, param, field, "Header")
		if err != nil {
			return err
		}

		op.AddParameter(param)
	}

	return nil
}
