package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateHeadersDoc(swagger *openapi3.T, op *openapi3.Operation, requestObjectType reflect.Type, fieldsFiltered schema.Fields) error {
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

		schema, err := b.schema.GenerateSchemaFor(swagger, field.Type, fieldsFiltered)
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
