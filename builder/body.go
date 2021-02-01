package builder

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

func (b *Builder) generateBodyDocumentation(op *openapi3.Operation, requestObjectType reflect.Type) error {
	bodyField, found := requestObjectType.FieldByName("Body")
	if found {
		body := openapi3.NewRequestBody()

		bodySchema, err := schemaFromType(bodyField.Type)
		if err != nil {
			return err
		}

		data, err := json.Marshal(bodySchema)
		if err != nil {
			panic(err)
		}
		fmt.Printf("schema %v : %s\n", bodyField.Type, data)

		bodyRef := &openapi3.RequestBodyRef{
			// Ref:   "#/components/requestBodies/pet",
			Value: body,
		}

		body.Content = openapi3.NewContentWithJSONSchema(bodySchema)
		op.RequestBody = bodyRef
	}

	return nil
}
