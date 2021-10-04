package builder

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/schmurfy/chipi/schema"
)

func (b *Builder) generateParametersDoc(r chi.Router, op *openapi3.Operation, requestObjectType reflect.Type, method string) error {
	pathField, found := requestObjectType.FieldByName("Path")
	if !found {
		return errors.New("wrong struct, Path field expected")
	}

	example, found := pathField.Tag.Lookup("example")
	if !found {
		return errors.New("missing tag `example`")
	}

	tctx := chi.NewRouteContext()
	if r.Match(tctx, method, example) {
		for _, key := range tctx.URLParams.Keys {
			// pathStruct must contain all defined keys
			paramField, found := pathField.Type.FieldByName(key)
			if !found {
				return errors.Errorf("wrong path struct, field %s expected", key)
			}

			schema, err := b.schema.GenerateSchemaFor(b.swagger, paramField.Type)
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
		return errors.Errorf("failed to match route: %s", example)
	}

	return nil
}

func prepareExample(t reflect.Type, val string) (interface{}, error) {
	var ex interface{}

	switch t.Kind() {
	case reflect.Slice, reflect.Struct, reflect.Map:
		ex = reflect.New(t).Interface()
		err := json.Unmarshal([]byte(val), &ex)
		if err != nil {
			return nil, errors.WithStack(err)
		}

	default:
		ex = val
	}

	return ex, nil
}

func fillParamFromTags(requestObjectType reflect.Type, param *openapi3.Parameter, f reflect.StructField, location string) error {
	var err error
	nilValue := reflect.New(requestObjectType)
	pathMethod, hasPathAnnotations := reflect.PtrTo(requestObjectType).MethodByName(fmt.Sprintf("CHIPI_%s_Annotations", location))

	// check for comments containing properties
	if hasPathAnnotations {
		ret := pathMethod.Func.Call([]reflect.Value{
			nilValue,
			reflect.ValueOf(param.Name),
		})

		if p, ok := ret[0].Interface().(*openapi3.Parameter); ok && (p != nil) {
			if p.Description != "" {
				param.Description = p.Description
			}

			if p.Example != nil {
				param.Example, err = prepareExample(f.Type, p.Example.(string))
				if err != nil {
					return err
				}

			}
		}
	}

	tag := schema.ParseJsonTag(f)

	if val, found := f.Tag.Lookup("example"); found {
		param.Example, err = prepareExample(f.Type, val)
		if err != nil {
			return err
		}
	}

	if tag.Description != nil {
		param.Description = *tag.Description
	}

	if tag.Style != nil {
		param.Style = *tag.Style
	}

	if tag.Explode != nil {
		param.Explode = tag.Explode
	}

	if tag.Deprecated != nil {
		param.Deprecated = *tag.Deprecated
	}

	if tag.Required != nil {
		param.Required = *tag.Required
	}

	return nil
}
