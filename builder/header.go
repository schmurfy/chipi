package builder

import (
	"context"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

func (b *Builder) generateHeadersDoc(ctx context.Context, swagger *openapi3.T, op *openapi3.Operation, requestObjectType reflect.Type) error {
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

		schema, err := b.schema.GenerateSchemaFor(ctx, swagger, field.Type)
		if err != nil {
			return err
		}

		name := field.Tag.Get("name")
		headerName := field.Name
		if name != "" {
			headerName = name
		}

		param := openapi3.NewHeaderParameter(headerName).
			WithSchema(schema.Value)

		err = fillParamFromTags(requestObjectType, param, field, "Header")
		if err != nil {
			return err
		}

		op.AddParameter(param)
	}

	return nil
}
