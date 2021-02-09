package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateResponseDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		description, found := responseField.Tag.Lookup("description")
		if found {
			resp.Description = &description
		}

		if responseField.Type.Kind() == reflect.Struct {
			responseSchema, err := schema.GenerateSchemaFor(b.swagger, responseField.Type)
			if err != nil {
				return err
			}

			resp.Content = openapi3.NewContentWithJSONSchemaRef(responseSchema)
		}

		op.Responses = make(openapi3.Responses)
		op.Responses["200"] = &openapi3.ResponseRef{
			Value: resp,
		}
	}

	return nil
}
