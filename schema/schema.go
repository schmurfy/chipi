package schema

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/shared"
)

var (
	_timeType = reflect.TypeOf(time.Time{})
)

type Schema struct {
}

func New() (*Schema, error) {
	return &Schema{}, nil
}

func (s *Schema) GenerateSchemaFor(ctx context.Context, doc *openapi3.T, t reflect.Type) (*openapi3.SchemaRef, error) {
	return s.generateSchemaFor(ctx, doc, t, 0, shared.AttributeInfo{}, shared.NewChipiCallbacks(nil))
}

func (s *Schema) GenerateFilteredSchemaFor(ctx context.Context, doc *openapi3.T, t reflect.Type, callbacksObject shared.ChipiCallbacks) (*openapi3.SchemaRef, error) {
	return s.generateSchemaFor(ctx, doc, t, 0, shared.AttributeInfo{}, callbacksObject)
}

func (s *Schema) generateSchemaFor(ctx context.Context, doc *openapi3.T, t reflect.Type, inlineLevel int, fieldInfo shared.AttributeInfo, callbacksObject shared.ChipiCallbacks) (*openapi3.SchemaRef, error) {
	fullName := typeName(t)

	if !fieldInfo.Empty() {
		filter, err := callbacksObject.FilterField(ctx, fieldInfo)
		if err != nil {
			return nil, err
		}

		if filter {
			return nil, nil
		}
	}

	schema := &openapi3.SchemaRef{}

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

		// []byte
		if t.Elem().Kind() == reflect.Uint8 {
			schema.Value = &openapi3.Schema{
				Type:   "string",
				Format: "binary",
			}

		} else {
			items, err := s.generateSchemaFor(ctx, doc, t.Elem(), 0, fieldInfo, callbacksObject)
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
		additionalProperties, err := s.generateSchemaFor(ctx, doc, t.Elem(), 0, fieldInfo, callbacksObject)
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
			return schema, nil
		}

		if doc.Components.Schemas == nil {
			doc.Components.Schemas = make(openapi3.Schemas)
		}

		// if we have an anonymous structure, inline it and stop there
		if (t.Name() == "") || (inlineLevel > 0) {
			var err error
			schema.Value, err = s.generateStructureSchema(ctx, doc, t, inlineLevel, fieldInfo, callbacksObject)
			return schema, err
		}

		// check if the structure already exists as component first
		_, found := doc.Components.Schemas[fullName]
		if !found {
			var err error
			ref := &openapi3.SchemaRef{}

			// forward declaration of the current type to handle recursion properly
			doc.Components.Schemas[fullName] = ref

			// fmt.Printf("%s - BEFORE: %+v\n", t.Name(), doc.Components.Schemas[t.Name()])
			ref.Value, err = s.generateStructureSchema(ctx, doc, t, inlineLevel, fieldInfo, callbacksObject)
			if err != nil {
				return nil, err
			}
			// fmt.Printf("%s - AFTER: %+v\n", t.Name(), doc.Components.Schemas[t.Name()])
		}

		schema.Ref = schemaReference(t)

	default:
		return nil, fmt.Errorf("unknown type: %v", t.Kind())
	}

	// Handle the case of enums
	if isEnum, enum := callbacksObject.EnumResolver(t); isEnum {
		_, found := doc.Components.Schemas[fullName]
		if !found {
			for _, enumEntry := range enum {
				schema.Value.OneOf = append(schema.Value.OneOf, &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Title: fmt.Sprint(enumEntry.Title),
						ExtensionProps: openapi3.ExtensionProps{
							Extensions: map[string]interface{}{
								"const": enumEntry.Value,
							},
						},
						Type:   "const",
						Format: schema.Value.Format,
					},
				})
			}
			doc.Components.Schemas[fullName] = &openapi3.SchemaRef{
				Value: schema.Value,
			}
		}

		schema.Ref = schemaReference(t)
		schema.Value = nil
	}

	return schema, nil
}

func typeName(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

func schemaReference(t reflect.Type) string {
	return fmt.Sprintf("#/components/schemas/%s", typeName(t))
}

func pkgName(t reflect.Type) string {
	parts := strings.Split(t.PkgPath(), "/")

	return parts[len(parts)-1]
}

func (s *Schema) generateStructureSchema(ctx context.Context, doc *openapi3.T, t reflect.Type, inlineLevel int, fieldInfo shared.AttributeInfo, callbacksObject shared.ChipiCallbacks) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	pkgName := shared.ToSnakeCase(pkgName(t))
	structName := shared.ToSnakeCase(t.Name())

	fieldInfo = fieldInfo.AppendPath(structName)

	filter, err := callbacksObject.FilterField(ctx, fieldInfo)
	if err != nil {
		return nil, err
	}

	if filter {
		return nil, nil
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := ParseJsonTag(f)

		if !f.IsExported() || (tag.Ignored != nil) && *tag.Ignored {
			continue
		}

		fieldName := shared.ToSnakeCase(f.Name)
		fi := fieldInfo.
			WithModelPath(pkgName + "." + structName + "." + fieldName).
			AppendPath(fieldName)

		fieldSchema, err := s.generateSchemaFor(ctx, doc, f.Type, inlineLevel-1, fi, callbacksObject)
		if err != nil {
			return nil, err
		}

		if fieldSchema == nil {
			continue
		}

		//Detect if field is anonymous, look into the schemas and use the same property
		if f.Anonymous && fieldSchema.Ref != "" && doc.Components.Schemas[f.Type.String()] != nil && doc.Components.Schemas[f.Type.String()].Value != nil {
			for name, property := range doc.Components.Schemas[f.Type.String()].Value.Properties {
				ret.WithPropertyRef(name, property)
			}

			//Ignore the anonymous field
			continue
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

				if tag.Description != nil {
					fieldSchema.Value.Description = *tag.Description
				}
			}

			// fmt.Printf("wtf: %s.%s (%s)\n", t.Name(), f.Name, fieldSchema.Ref)
			// fieldSchema.Value = openapi3.NewSchema()
		} else {
			fieldSchema.Value.ReadOnly = (tag.ReadOnly != nil) && *tag.ReadOnly
			fieldSchema.Value.Nullable = (tag.Nullable != nil) && *tag.Nullable
			fieldSchema.Value.Deprecated = (tag.Deprecated != nil) && *tag.Deprecated

			if tag.Description != nil {
				fieldSchema.Value.Description = *tag.Description
			}

			if tag.Example != nil {
				fieldSchema.Value.Example = *tag.Example
			}

			if tag.Required != nil && *tag.Required {
				ret.Required = append(fieldSchema.Value.Required, fieldName)
			}
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
