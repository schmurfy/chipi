package builder

import (
	"strings"

	"context"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ruwanego/chipi/schema"
	"github.com/ruwanego/chipi/shared"
	"github.com/ruwanego/chipi/wrapper"
)

func (b *Builder) generateResponseDoc(ctx context.Context, swagger *openapi3.T, op *openapi3.Operation, requestObject interface{}, requestObjectType reflect.Type, callbacksObject shared.ChipiCallbacks) error {
	responses := openapi3.Responses{}

	responseField, found := requestObjectType.FieldByName("Response")
	if found {
		resp := openapi3.NewResponse()

		// check that a body decoder is available
		if _, ok := requestObject.(wrapper.ResponseEncoder); !ok {
			return fmt.Errorf("%s must implement ResponseEncoder", requestObjectType.Name())
		}

		contentTypeStr, hasContentType := responseField.Tag.Lookup("content-type")
		if !hasContentType {
			contentTypeStr = "application/json"
		}

		typ := responseField.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		err := fillResponseFromTags(requestObjectType, resp, responseField)
		if err != nil {
			return err
		}

		resp.Content = openapi3.NewContent()

		if typ.Kind() == reflect.Struct {
			responseSchema, err := b.schema.GenerateFilteredSchemaFor(ctx, swagger, typ, callbacksObject)
			if err != nil {
				return err
			}

			for _, contentType := range strings.Split(contentTypeStr, ",") {
				resp.Content[strings.TrimSpace(contentType)] = &openapi3.MediaType{
					Schema: responseSchema,
				}
			}
		} else if typ.Kind() == reflect.Slice {
			responseSchema, err := b.schema.GenerateFilteredSchemaFor(ctx, swagger, typ, callbacksObject)
			if err != nil {
				return err
			}

			contentTypes := strings.Split(contentTypeStr, ",")
			if responseSchema.Value.Format == "binary" && !hasContentType {
				contentTypes = []string{"application/octet-stream"}
			}

			for _, contentType := range contentTypes {
				resp.Content[strings.TrimSpace(contentType)] = &openapi3.MediaType{
					Schema: responseSchema,
				}
			}
		}

		responses.Set("200", &openapi3.ResponseRef{
			Value: resp,
		})
	} else {
		// if no response provided generate a default 204 code response
		noData := "no data"
		responses.Set("204", &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &noData,
			},
		})
	}

	// add endpoint-specific errors
	errorsField, found := requestObjectType.FieldByName("Errors")
	if found {
		errorsType := errorsField.Type
		if errorsType.Kind() == reflect.Ptr {
			errorsType = errorsType.Elem()
		}

		if errorsType.Kind() == reflect.Struct {
			for i := 0; i < errorsType.NumField(); i++ {
				f := errorsType.Field(i)

				statusCode, ok := f.Tag.Lookup("chipi")
				if !ok {
					continue
				}

				contentTypeStr, ok := f.Tag.Lookup("content-type")
				if !ok {
					contentTypeStr = "application/json"
				}

				resp := openapi3.NewResponse()

				tag := schema.ParseJsonTag(f)
				if tag.Description != nil {
					resp.Description = tag.Description
				} else {
					desc := fmt.Sprintf("Error %s", statusCode)
					resp.Description = &desc
				}

				fieldType := f.Type
				if fieldType.Kind() == reflect.Ptr {
					fieldType = fieldType.Elem()
				}

				// We allow omitting the model by using an empty struct or just interface{}?
				// To keep it simple, if it's not a struct, we don't generate schema unless it's known
				// But let's just generate schema for it like we do for Response.
				responseSchema, err := b.schema.GenerateFilteredSchemaFor(ctx, swagger, fieldType, callbacksObject)
				if err != nil {
					return err
				}

				resp.Content = openapi3.NewContent()
				for _, contentType := range strings.Split(contentTypeStr, ",") {
					resp.Content[strings.TrimSpace(contentType)] = &openapi3.MediaType{
						Schema: responseSchema,
					}
				}

				responses.Set(statusCode, &openapi3.ResponseRef{
					Value: resp,
				})
			}
		}
	}

	// add global responses
	if len(b.globalErrors) > 0 {
		for _, gErr := range b.globalErrors {
			codeStr := fmt.Sprintf("%d", gErr.StatusCode)
			if responses.Value(codeStr) != nil {
				continue // Do not override if already defined (e.g. by endpoint-specific Errors)
			}

			resp := openapi3.NewResponse()

			description := gErr.Description
			if description != "" {
				resp.Description = &description
			} else {
				noDesc := "Global Error Response"
				resp.Description = &noDesc
			}

			if gErr.Model != nil {
				gErrType := reflect.TypeOf(gErr.Model)
				if gErrType.Kind() == reflect.Ptr {
					gErrType = gErrType.Elem()
				}

				responseSchema, err := b.schema.GenerateFilteredSchemaFor(ctx, swagger, gErrType, callbacksObject)
				if err != nil {
					return err
				}

				resp.Content = openapi3.NewContent()
				for _, ct := range strings.Split(gErr.ContentType, ",") {
					resp.Content[strings.TrimSpace(ct)] = &openapi3.MediaType{
						Schema: responseSchema,
					}
				}
			}

			responses.Set(codeStr, &openapi3.ResponseRef{
				Value: resp,
			})
		}
	}

	op.Responses = &responses

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
