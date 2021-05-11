package builder

import (
	"errors"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/schmurfy/chipi/schema"
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

		schema, err := schema.GenerateSchemaFor(b.swagger, field.Type)
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
