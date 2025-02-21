package schema

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/shared"
)

var (
	_timeType         = reflect.TypeOf(time.Time{})
	_genericTypeRegex = regexp.MustCompile(`(.+)\[(.+)\]`)
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

func (s *Schema) newGenerateSchemaCallback(ctx context.Context, doc *openapi3.T, inlineLevel int, callbacksObject shared.ChipiCallbacks) shared.GenerateSchemaCallbackType {
	return func(t reflect.Type, fieldInfo shared.AttributeInfo) (*openapi3.SchemaRef, error) {
		return s.generateSchemaFor(ctx, doc, t, inlineLevel, fieldInfo, callbacksObject)
	}
}

func (s *Schema) generateSchemaWithCastedType(ctx context.Context, doc *openapi3.T, t reflect.Type, castName string, inlineLevel int, fieldInfo shared.AttributeInfo, callbacksObject shared.ChipiCallbacks) (*openapi3.SchemaRef, error) {
	if !fieldInfo.Empty() {
		filter, err := callbacksObject.FilterField(ctx, fieldInfo)
		if err != nil {
			return nil, err
		}

		if filter {
			return nil, nil
		}
	}

	if doc.Components.Schemas == nil {
		doc.Components.Schemas = make(openapi3.Schemas)
	}

	customName := fmt.Sprintf("custom_%s", castName)
	if _, found := doc.Components.Schemas[customName]; !found {
		var createRef bool
		schema := &openapi3.SchemaRef{}
		schema.Value, createRef = callbacksObject.SchemaResolver(fieldInfo, castName, t, s.newGenerateSchemaCallback(ctx, doc, inlineLevel, callbacksObject))
		if createRef {
			doc.Components.Schemas[customName] = schema
		} else {
			return schema, nil
		}
	}

	return &openapi3.SchemaRef{
		Ref: fmt.Sprintf("#/components/schemas/%s", customName),
	}, nil
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
			Type: "object",
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: additionalProperties,
			},
		}

	// struct schemas should be stored as components
	case reflect.Struct:
		if t == _timeType {
			schema.Value = openapi3.NewDateTimeSchema()
			return schema, nil
		}

		if doc.Components == nil {
			doc.Components = &openapi3.Components{}
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

			ref.Value, err = s.generateStructureSchema(ctx, doc, t, inlineLevel, fieldInfo, callbacksObject)
			if err != nil {
				return nil, err
			}
		}

		schema.Ref = schemaReference(t)
	case reflect.Interface:
		schema.Value = openapi3.NewSchema()

	default:
		return nil, fmt.Errorf("unknown type: %v", t.Kind())
	}

	// Handle the case of enums
	if isEnum, enum := callbacksObject.EnumResolver(t); isEnum {
		_, found := doc.Components.Schemas[fullName]
		if !found {
			enumDescriptions := []any{}
			for _, enumEntry := range enum {
				schema.Value.Enum = append(schema.Value.Enum, enumEntry.Value)
				enumDescriptions = append(enumDescriptions, enumEntry.Title)
			}
			if len(enumDescriptions) > 0 {
				schema.Value.Extensions = map[string]any{
					"x-enum-varnames": enumDescriptions,
				}
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

	typeName := t.String()

	// When using generic structures reflect's type name looks something like
	// module.genericStruct[full/module/path/module.subType]
	// OpenAPI does not support "[]" or "/", so we replace this by module.genericStruct..module.subType
	match := _genericTypeRegex.FindStringSubmatch(typeName)
	if match == nil {
		return typeName
	} else {
		return fmt.Sprintf("%s..%s", match[1], pkgName(match[2]))
	}
}

func schemaReference(t reflect.Type) string {
	return fmt.Sprintf("#/components/schemas/%s", typeName(t))
}

func pkgName(p string) string {
	parts := strings.Split(p, "/")
	return parts[len(parts)-1]
}

func typePkgName(t reflect.Type) string {
	return pkgName(t.PkgPath())
}

func (s *Schema) generateStructureSchema(ctx context.Context, doc *openapi3.T, t reflect.Type, inlineLevel int, fieldInfo shared.AttributeInfo, callbacksObject shared.ChipiCallbacks) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	pkgName := shared.ToSnakeCase(typePkgName(t))
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
		fTypeName := typeName(f.Type)
		tag := ParseJsonTag(f)

		if !f.IsExported() || (tag.Ignored != nil) && *tag.Ignored {
			continue
		}

		fieldName := shared.ToSnakeCase(f.Name)
		fi := fieldInfo.
			WithModelPath(pkgName + "." + structName + "." + fieldName).
			AppendPath(fieldName)

		var fieldSchema *openapi3.SchemaRef
		if tag.CastName != nil {
			fieldSchema, err = s.generateSchemaWithCastedType(ctx, doc, f.Type, *tag.CastName, inlineLevel-1, fi, callbacksObject)
		} else {
			fieldSchema, err = s.generateSchemaFor(ctx, doc, f.Type, inlineLevel-1, fi, callbacksObject)
		}

		if err != nil {
			return nil, err
		}

		if fieldSchema == nil {
			continue
		}

		//Detect if field is anonymous, look into the schemas and use the same property
		if f.Anonymous && fieldSchema.Ref != "" && doc.Components.Schemas[fTypeName] != nil && doc.Components.Schemas[fTypeName].Value != nil {
			for name, property := range doc.Components.Schemas[fTypeName].Value.Properties {
				ret.WithPropertyRef(name, property)
			}

			//Ignore the anonymous field
			continue
		}

		if fieldSchema.Ref != "" {
			if tag.ReadOnly != nil || tag.Nullable != nil || tag.Deprecated != nil || tag.Example != nil {
				// As those tags are not supported when using a ref we need to use a allOf
				// with the ref inside and apply the tag to the parent
				if fieldSchema.Value == nil {
					fieldSchema.Value = openapi3.NewSchema()
				}

				fieldSchema.Value.AllOf = openapi3.SchemaRefs{openapi3.NewSchemaRef(fieldSchema.Ref, nil)}
				fieldSchema.Ref = ""
				fieldSchema.Value.ReadOnly = tag.GetReadOnly()
				fieldSchema.Value.Nullable = tag.GetNullable()
				fieldSchema.Value.Deprecated = tag.GetDeprecated()
				if tag.Example != nil {
					fieldSchema.Value.Example = tag.GetExample()
				}
				if tag.Description != nil {
					fieldSchema.Value.Description = tag.GetDescription()
				}
			} else if tag.Description != nil {
				fieldSchema.Value = openapi3.NewSchema()
				fieldSchema.Value.Description = tag.GetDescription()
			}
		} else {
			if fieldSchema.Value == nil {
				fieldSchema.Value = openapi3.NewSchema()
			}
			fieldSchema.Value.ReadOnly = tag.GetReadOnly()
			fieldSchema.Value.Nullable = tag.GetNullable()
			fieldSchema.Value.Deprecated = tag.GetDeprecated()

			if tag.Description != nil {
				fieldSchema.Value.Description = tag.GetDescription()
			}

			if tag.Example != nil {
				fieldSchema.Value.Example = tag.GetExample()
			}
		}

		if tag.Required != nil && *tag.Required {
			ret.Required = append(ret.Required, fieldName)
		}
		ret.WithPropertyRef(tag.Name, fieldSchema)
	}

	// object type requires properties
	if ret.Properties != nil {
		ret.Type = "object"
	}

	return ret, nil
}
