package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

func (b *Builder) generateResponseDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		description, found := responseField.Tag.Lookup("description")
		if found {
			resp.Description = &description
		}

		contentType, hasContentType := responseField.Tag.Lookup("content-type")
		if !hasContentType {
			contentType = "application/json"
		}

		typ := responseField.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() == reflect.Struct {
			responseSchema, err := b.schema.GenerateSchemaFor(b.swagger, typ)
			if err != nil {
				return err
			}

			resp.Content = openapi3.Content{
				contentType: &openapi3.MediaType{
					Schema: responseSchema,
				},
			}
		}

		op.Responses = make(openapi3.Responses)
		op.Responses["200"] = &openapi3.ResponseRef{
			Value: resp,
		}
	}

	return nil
}
