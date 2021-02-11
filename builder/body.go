package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateBodyDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	bodyField, found := requestObjectType.FieldByName("Body")
	if found {
		bodySchema, err := schema.GenerateSchemaFor(b.swagger, bodyField.Type)
		if err != nil {
			return err
		}

		contentType, found := bodyField.Tag.Lookup("content-type")
		if !found {
			contentType = "application/json"
		}

		body := openapi3.NewRequestBody()
		bodyRef := &openapi3.RequestBodyRef{Value: body}

		body.Content = openapi3.Content{
			contentType: &openapi3.MediaType{
				Schema: bodySchema,
			},
		}

		if val, found := bodyField.Tag.Lookup("description"); found {
			body.Description = val
		}

		if val, found := bodyField.Tag.Lookup("required"); found {
			body.Required = (val == "true")
		}

		op.RequestBody = bodyRef
	}

	return nil
}
