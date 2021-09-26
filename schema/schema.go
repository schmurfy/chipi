package schema

import (
	"fmt"
	"reflect"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	_timeType = reflect.TypeOf(time.Time{})
)

type Schema struct {
	cache map[string]*openapi3.SchemaRef
}

func New() (*Schema, error) {
	return &Schema{
		cache: map[string]*openapi3.SchemaRef{},
	}, nil
}

func (s *Schema) GenerateSchemaFor(doc *openapi3.T, t reflect.Type) (schema *openapi3.SchemaRef, err error) {
	cached, found := s.cache[t.String()]
	if found {
		return cached, nil
	}

	schema = &openapi3.SchemaRef{}
	s.cache[t.String()] = schema

	defer func() {
		if err != nil {
			delete(s.cache, t.String())
		}
	}()

	// test pointed value for pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {

	// basic types
	case reflect.String:
		schema.Value = openapi3.NewStringSchema()

	case reflect.Bool:
		schema.Value = openapi3.NewBoolSchema()

	case reflect.Int8, reflect.Int16, reflect.Int32:
		schema.Value = openapi3.NewInt32Schema()

	case reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Value = openapi3.NewInt64Schema()

	case reflect.Float32, reflect.Float64:
		schema.Value = &openapi3.Schema{
			Type:   "number",
			Format: "double",
		}

	// complex types
	case reflect.Slice:
		var items *openapi3.SchemaRef

		// []byte
		if t.Elem().Kind() == reflect.Uint8 {
			schema.Value = &openapi3.Schema{
				Type:   "string",
				Format: "binary",
			}

		} else {
			items, err = s.GenerateSchemaFor(doc, t.Elem())
			if err != nil {
				return nil, err
			}

			schema.Value = &openapi3.Schema{
				Type:  "array",
				Items: items,
			}

		}

	case reflect.Map:
		// schema.Type = "object"
		// additionalProperties, err := g.generateSchemaRefFor(parents, t.Elem())
		// if err != nil {
		// 	return nil, err
		// }
		// if additionalProperties != nil {
		// 	g.SchemaRefs[additionalProperties]++
		// 	schema.AdditionalProperties = additionalProperties
		// }

	// struct schemas should be stored as components
	case reflect.Struct:
		if t == _timeType {
			schema.Value = openapi3.NewDateTimeSchema()
			return
		}

		// if we have an anonymous structure, inline it and stop there
		if t.Name() == "" {
			schema.Value, err = s.generateStructureSchema(doc, t)
			// return wether an error occured ot not so don't bother
			// checking err which will be returned anyway
			return
		}

		if doc.Components.Schemas == nil {
			doc.Components.Schemas = make(openapi3.Schemas)
		}

		// check if the structure already exists as component first
		_, found := doc.Components.Schemas[t.Name()]
		if !found {
			var sch *openapi3.Schema
			sch, err = s.generateStructureSchema(doc, t)
			if err != nil {
				return
			}

			// register the schema
			doc.Components.Schemas[t.Name()] = &openapi3.SchemaRef{Value: sch}
		}

		schema.Ref = structReference(t)
	}

	return
}

func structReference(t reflect.Type) string {
	return fmt.Sprintf("#/components/schemas/%s", t.Name())
}

func (s *Schema) generateStructureSchema(doc *openapi3.T, t reflect.Type) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := ParseJsonTag(f)

		if (tag.Ignored != nil) && *tag.Ignored {
			continue
		}

		fieldSchema, err := s.GenerateSchemaFor(doc, f.Type)
		if err != nil {
			return nil, err
		}

		if fieldSchema.Ref != "" {
			if (tag.Nullable != nil) && *tag.Nullable {
				// nullableSchemaRef := openapi3.NewSchemaRef("", &openapi3.Schema{
				// 	OneOf: openapi3.SchemaRefs{
				// 		openapi3.NewSchemaRef("", &openapi3.Schema{Type: "null"}),
				// 		fieldSchema,
				// 	},
				// })

				// fieldSchema = nullableSchemaRef

				fieldSchema.Value = openapi3.NewSchema()
				fieldSchema.Value.Nullable = *tag.Nullable
			}

			// fmt.Printf("wtf: %s.%s (%s)\n", t.Name(), f.Name, fieldSchema.Ref)
			// fieldSchema.Value = openapi3.NewSchema()
		} else {
			if (tag.ReadOnly != nil) && *tag.ReadOnly {
				fieldSchema.Value.ReadOnly = *tag.ReadOnly
			}

			if (tag.Nullable != nil) && *tag.Nullable {
				fieldSchema.Value.Nullable = *tag.Nullable
			}

		}

		ret.WithPropertyRef(tag.Name, fieldSchema)
	}

	// object type requires properties
	if ret.Properties != nil {
		ret.Type = "object"
	}

	return ret, nil
}
