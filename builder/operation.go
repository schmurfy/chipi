package builder

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

func generateOperationDoc(op *openapi3.Operation, requestObjectType reflect.Type) error {
	err := fillOperationFromComments(requestObjectType, op)
	if err != nil {
		return err
	}

	return nil
}

func fillOperationFromComments(requestObjectType reflect.Type, op *openapi3.Operation) error {
	nilValue := reflect.New(requestObjectType)

	opMethod, hasOperationAnnotations := reflect.PtrTo(requestObjectType).MethodByName("CHIPI_Operation_Annotations")
	if hasOperationAnnotations {
		ret := opMethod.Func.Call([]reflect.Value{
			nilValue,
		})

		if o, ok := ret[0].Interface().(*openapi3.Operation); ok && (o != nil) {
			if len(o.Tags) > 0 {
				op.Tags = append(op.Tags, o.Tags...)
			}

			if o.Summary != "" {
				op.Summary = o.Summary
			}

			if o.Description != "" {
				op.Description = o.Description
			}

			op.Deprecated = o.Deprecated
		}
	}

	return nil
}
