package builder

import (
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/schema"
	"github.com/schmurfy/chipi/wrapper"
)

func (b *Builder) generateResponseDoc(swagger *openapi3.T, op *openapi3.Operation, requestObject interface{}, requestObjectType reflect.Type) error {
	responses := make(openapi3.Responses)

	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		// check that a body decoder is available
		if _, ok := requestObject.(wrapper.ResponseEncoder); !ok {
			return fmt.Errorf("%s must implement ResponseEncoder", requestObjectType.Name())
		}

		contentType, hasContentType := responseField.Tag.Lookup("content-type")
		if !hasContentType {
			contentType = "application/json"
		}

		typ := responseField.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		err := fillResponseFromTags(requestObjectType, resp, responseField)
		if err != nil {
			return err
		}

		if typ.Kind() == reflect.Struct {
			responseSchema, err := b.schema.GenerateSchemaFor(swagger, typ)
			if err != nil {
				return err
			}

			resp.Content = openapi3.Content{
				contentType: &openapi3.MediaType{
					Schema: responseSchema,
				},
			}
		}

		responses["200"] = &openapi3.ResponseRef{
			Value: resp,
		}
	} else {
		// if no response provided generate a default 204 code response
		responses["204"] = &openapi3.ResponseRef{}
	}

	op.Responses = responses

	return nil
}

func fillResponseFromTags(requestObjectType reflect.Type, resp *openapi3.Response, f reflect.StructField) error {
	nilValue := reflect.New(requestObjectType)

	opMethod, hasOperationAnnotations := reflect.PtrTo(requestObjectType).MethodByName("CHIPI_Response_Annotations")
	if hasOperationAnnotations {
		ret := opMethod.Func.Call([]reflect.Value{
			nilValue,
			reflect.ValueOf(""),
		})

		if p, ok := ret[0].Interface().(*openapi3.Parameter); ok && (p != nil) {
			if p.Description != "" {
				resp.Description = &p.Description
			}
		}
	}

	tag := schema.ParseJsonTag(f)

	if tag.Description != nil {
		resp.Description = tag.Description
	}

	return nil
}
