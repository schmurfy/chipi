package builder

import (
	"encoding/json"
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

			err = fillParamFromTags(param, paramField)
			if err != nil {
				return err
			}

			op.AddParameter(param)
		}
	} else {
		return fmt.Errorf("failed to match route: %s", example)
	}

	return nil
}

func fillParamFromTags(param *openapi3.Parameter, f reflect.StructField) error {
	if val, found := f.Tag.Lookup("example"); found {
		if f.Type.Kind() == reflect.Slice {
			ex := reflect.New(f.Type).Interface()

			err := json.Unmarshal([]byte(val), &ex)
			if err != nil {
				return err
			}
			param.Example = ex
		} else {
			param.Example = val
		}
	}

	if val, found := f.Tag.Lookup("description"); found {
		param.Description = val
	}

	if val, found := f.Tag.Lookup("deprecated"); found {
		param.Deprecated = (val == "true")
	}

	if val, found := f.Tag.Lookup("style"); found {
		param.Style = val
	}

	if val, found := f.Tag.Lookup("explode"); found {
		b := (val == "true")
		param.Explode = &b
	}

	return nil
}
