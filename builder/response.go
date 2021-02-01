package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

func (b *Builder) generateResponseDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		responseSchema, err := schemaFromType(responseField.Type)
		if err != nil {
			return err
		}

		resp.Content = openapi3.NewContentWithJSONSchema(responseSchema)

		op.Responses = make(openapi3.Responses)
		op.Responses["200"] = &openapi3.ResponseRef{
			Value: resp,
		}
	}

	return nil
}
