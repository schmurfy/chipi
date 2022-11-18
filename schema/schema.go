package schema

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/slices"
)

var (
	_timeType = reflect.TypeOf(time.Time{})
)

type Schema struct {
}

type Fields struct {
	Protected []string
	Whitelist []string
}

func New() (*Schema, error) {
	return &Schema{}, nil
}

func (s *Schema) GenerateSchemaFor(doc *openapi3.T, t reflect.Type, fieldsFiltered Fields) (schema *openapi3.SchemaRef, err error) {
	// return s.generateSchemaFor(doc, t, 1)
	return s.generateSchemaFor(doc, t, fieldsFiltered, 0)
}

func (s *Schema) generateSchemaFor(doc *openapi3.T, t reflect.Type, fieldsFiltered Fields, inlineLevel int) (schema *openapi3.SchemaRef, err error) {
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

	case reflect.Int, reflect.Int64, reflect.Uint,
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
			items, err = s.generateSchemaFor(doc, t.Elem(), fieldsFiltered, 0)
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
		additionalProperties, err := s.generateSchemaFor(doc, t.Elem(), fieldsFiltered, 0)
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
			schema.Value, err = s.generateStructureSchema(doc, t, fieldsFiltered, inlineLevel)

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
			ref.Value, err = s.generateStructureSchema(doc, t, fieldsFiltered, inlineLevel)
			if err != nil {
				return
			}
			// fmt.Printf("%s - AFTER: %+v\n", t.Name(), doc.Components.Schemas[t.Name()])
		}

		schema.Ref = structReference(t)

	default:
		return nil, fmt.Errorf("unknwon type: %v", t.Kind())
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

func HasSharedValues[T comparable](l1 []T, l2 []T) bool {
	for _, v1 := range l1 {
		for _, v2 := range l2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

func (s *Schema) generateStructureSchema(doc *openapi3.T, t reflect.Type, fieldsFiltered Fields, inlineLevel int) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	toDelete := make([]string, 0)
	enforceWhitelist := false

	packages := strings.Split(t.String(), ".")

	for i := 0; i < t.NumField(); i++ {
		if len(packages) != 2 {
			continue
		}
		f := t.Field(i)
		fieldName := fmt.Sprintf("%s.%s.%s", strings.ToLower(packages[0]), strings.ToLower(packages[1]), strings.ToLower(f.Name))

		if slices.Contains(fieldsFiltered.Whitelist, fieldName) {
			enforceWhitelist = true
		} else {
			toDelete = append(toDelete, fieldName)
		}
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := ParseJsonTag(f)

		//Create registry string
		if len(packages) == 2 {
			field := fmt.Sprintf("%s.%s.%s", strings.ToLower(packages[0]), strings.ToLower(packages[1]), strings.ToLower(f.Name))
			if !enforceWhitelist && len(fieldsFiltered.Protected) > 0 && slices.Contains(fieldsFiltered.Protected, field) {
				continue
			} else if enforceWhitelist && slices.Contains(toDelete, field) {
				continue
			}
		}

		if (tag.Ignored != nil) && *tag.Ignored {
			continue
		}

		fieldSchema, err := s.generateSchemaFor(doc, f.Type, fieldsFiltered, inlineLevel-1)
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
