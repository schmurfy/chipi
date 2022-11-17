package builder

import (
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateQueryParametersDoc(swagger *openapi3.T, op *openapi3.Operation, requestObjectType reflect.Type, fieldsFiltered schema.Fields) error {
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

		fieldSchema, err := b.schema.GenerateSchemaFor(swagger, field.Type, fieldsFiltered)
		if err != nil {
			return err
		}

		param := openapi3.NewQueryParameter(strings.ToLower(field.Name))

		if (fieldSchema.Ref != "") || (fieldSchema.Value.Type == "object") {
			// we need to wrap the schema
			param.Content = openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: fieldSchema,
				},
			}
		} else {
			// just add the schema directly for basic types
			param = param.WithSchema(fieldSchema.Value)
		}

		err = fillParamFromTags(requestObjectType, param, field, "Query")
		if err != nil {
			return err
		}

		op.AddParameter(param)
	}

	return nil
}
