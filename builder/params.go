package builder

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateParametersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type, method string) error {
	pathField, found := requestObjectType.FieldByName("Path")
	if !found {
		return errors.New("wrong struct, Path field expected")
	}

	// example := pathField.Tag.Get("example")
	example, found := pathField.Tag.Lookup("example")
	if !found {
		return fmt.Errorf("missing tag `example`")
	}

	tctx := chi.NewRouteContext()
	if r.Match(tctx, method, example) {

		for _, key := range tctx.URLParams.Keys {
			// pathStruct must contain all defined keys
			paramField, found := pathField.Type.FieldByName(key)
			if !found {
				return fmt.Errorf("wrong path struct, field %s expected", key)
			}

			schema, err := schema.GenerateSchemaFor(b.swagger, paramField.Type)
			if err != nil {
				return err
			}

			param := openapi3.NewPathParameter(key).
				WithSchema(schema.Value)

			paramsExample, found := paramField.Tag.Lookup("example")
			if found {
				param.Example = paramsExample
			}

			op.AddParameter(param)
		}
	} else {
		return fmt.Errorf("failed to match route: %s", example)
	}

	return nil
}
