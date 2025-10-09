package builder

import (
	"context"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/schmurfy/chipi/schema"
	"github.com/schmurfy/chipi/shared"
)

func (b *Builder) generateQueryParametersDoc(ctx context.Context, swagger *openapi3.T, op *openapi3.Operation, requestObjectType reflect.Type, callbacksObject shared.ChipiCallbacks) error {
	pathField, found := requestObjectType.FieldByName("Query")
	if !found {
		return nil
	}

	queryStructType := pathField.Type
	if queryStructType.Kind() != reflect.Struct {
		return errors.Errorf("expected struct for Query : %s ", requestObjectType.Name())
	}

	for _, field := range reflect.VisibleFields(queryStructType) {
		if !field.IsExported() || field.Anonymous {
			continue
		}

		fieldSchema, err := b.schema.GenerateSchemaFor(ctx, swagger, field.Type)
		if err != nil {
			return err
		}
		parsedTag := schema.ParseJsonTag(field)

		if parsedTag.Ignored != nil && *parsedTag.Ignored {
			continue
		}

		name := parsedTag.Name
		if name == field.Name {
			name = shared.ToSnakeCase(field.Name)
		}

		if parsedTag.CastName != nil {
			fieldSchema.Value, _ = callbacksObject.SchemaResolver(shared.AttributeInfo{}, *parsedTag.CastName, field.Type, func(t reflect.Type, fieldInfo shared.AttributeInfo) (*openapi3.SchemaRef, error) {
				return fieldSchema, nil
			})
			op.AddParameter(openapi3.NewQueryParameter(name).WithSchema(fieldSchema.Value))
			continue
		}

		param := openapi3.NewQueryParameter(name)

		if (fieldSchema.Ref != "") || (fieldSchema.Value.Type.Includes(openapi3.TypeObject)) {
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
