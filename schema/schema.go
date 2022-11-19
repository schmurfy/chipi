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
}

func New() (*Schema, error) {
	return &Schema{}, nil
}

func (s *Schema) GenerateSchemaFor(doc *openapi3.T, t reflect.Type) (schema *openapi3.SchemaRef, err error) {
	// return s.generateSchemaFor(doc, t, 1)
	return s.generateSchemaFor(doc, t, 0)
}

func (s *Schema) generateSchemaFor(doc *openapi3.T, t reflect.Type, inlineLevel int) (schema *openapi3.SchemaRef, err error) {
	fullName := typeName(t)

	if doc.Components.Schemas != nil {
		cached, found := doc.Components.Schemas[fullName]
		if found {
			return cached, nil
		}
	}

	schema = &openapi3.SchemaRef{}

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
			items, err = s.generateSchemaFor(doc, t.Elem(), 0)
			if err != nil {
				return nil, err
			}

			// if (items.Ref == "") && (items.Value == nil) {
			// 	return nil, errors.Errorf("invalid schema for %s", t.Elem().String())
			// }

			// fmt.Printf("items: %+v\n", items)

			schema.Value = &openapi3.Schema{
				Type:  "array",
				Items: items,
			}

		}

	case reflect.Map:
		additionalProperties, err := s.generateSchemaFor(doc, t.Elem(), 0)
		if err != nil {
			return nil, err
		}

		schema.Value = &openapi3.Schema{
			Type:                 "object",
			AdditionalProperties: additionalProperties,
		}

	// struct schemas should be stored as components
	case reflect.Struct:
		if t == _timeType {
			schema.Value = openapi3.NewDateTimeSchema()
			return
		}

		if doc.Components.Schemas == nil {
			doc.Components.Schemas = make(openapi3.Schemas)
		}

		// if we have an anonymous structure, inline it and stop there
		if (t.Name() == "") || (inlineLevel > 0) {
			schema.Value, err = s.generateStructureSchema(doc, t, inlineLevel)
			// return wether an error occurred ot not so don't bother
			// checking err which will be returned anyway
			return
		}

		// check if the structure already exists as component first
		_, found := doc.Components.Schemas[fullName]
		if !found {
			ref := &openapi3.SchemaRef{}

			// forward declaration of the current type to handle recursion properly
			doc.Components.Schemas[fullName] = ref

			// fmt.Printf("%s - BEFORE: %+v\n", t.Name(), doc.Components.Schemas[t.Name()])
			ref.Value, err = s.generateStructureSchema(doc, t, inlineLevel)
			if err != nil {
				return
			}
			// fmt.Printf("%s - AFTER: %+v\n", t.Name(), doc.Components.Schemas[t.Name()])
		}

		schema.Ref = structReference(t)

	default:
		return nil, fmt.Errorf("unknown type: %v", t.Kind())
	}

	return
}

func typeName(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

func structReference(t reflect.Type) string {
	return fmt.Sprintf("#/components/schemas/%s", typeName(t))
}

func (s *Schema) generateStructureSchema(doc *openapi3.T, t reflect.Type, inlineLevel int) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := ParseJsonTag(f)

		if (tag.Ignored != nil) && *tag.Ignored {
			continue
		}

		fieldSchema, err := s.generateSchemaFor(doc, f.Type, inlineLevel-1)
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
			fieldSchema.Value.ReadOnly = (tag.ReadOnly != nil) && *tag.ReadOnly
			fieldSchema.Value.Nullable = (tag.Nullable != nil) && *tag.Nullable
			fieldSchema.Value.Deprecated = (tag.Deprecated != nil) && *tag.Deprecated

			// if f.Name == "Coordinates" {
			// 	fmt.Printf("[DD] %s.%s : %+v\n", t.Name(), f.Name, tag)
			// }

		}

		ret.WithPropertyRef(tag.Name, fieldSchema)
	}

	// object type requires properties
	if ret.Properties != nil {
		ret.Type = "object"
	}

	return ret, nil
}
