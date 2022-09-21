package builder

import (
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/schema"
	"github.com/schmurfy/chipi/wrapper"
)

func (b *Builder) generateBodyDoc(op *openapi3.Operation, requestObject interface{}, requestObjectType reflect.Type) error {
	bodyField, found := requestObjectType.FieldByName("Body")
	if found {
		bodySchema, err := b.schema.GenerateSchemaFor(b.swagger, bodyField.Type)
		if err != nil {
			return err
		}

		// check that a body decoder is available
		if _, ok := requestObject.(wrapper.BodyDecoder); !ok {
			return fmt.Errorf("%s must implement BodyDecoder", requestObjectType.Name())
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

		tag := schema.ParseJsonTag(bodyField)

		if tag.Description != nil {
			body.Description = *tag.Description
		}

		if tag.Required != nil {
			body.Required = *tag.Required
		}

		op.RequestBody = bodyRef
	}

	return nil
}
