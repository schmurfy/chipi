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

			err = fillParamFromTags(requestObjectType, param, paramField, "Path")
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

func fillParamFromTags(requestObjectType reflect.Type, param *openapi3.Parameter, f reflect.StructField, location string) error {
	nilValue := reflect.New(requestObjectType)
	pathMethod, hasPathAnnotations := reflect.PtrTo(requestObjectType).MethodByName(fmt.Sprintf("Chipi_%s_Annotations", location))

	// check for comments containing properties
	if hasPathAnnotations {
		ret := pathMethod.Func.Call([]reflect.Value{
			nilValue,
			reflect.ValueOf(param.Name),
		})

		if p, ok := ret[0].Interface().(*openapi3.Parameter); ok {
			if p.Description != "" {
				param.Description = p.Description
			}

			if p.Example != nil {
				param.Example = p.Example
			}
		}
	}

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
