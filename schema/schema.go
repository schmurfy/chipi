package schema

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	_timeType = reflect.TypeOf(time.Time{})
)

func GenerateSchemaFor(doc *openapi3.Swagger, t reflect.Type) (schema *openapi3.SchemaRef, err error) {
	// Get TypeInfo
	// typeInfo := jsoninfo.GetTypeInfo(t)

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
		items, err := GenerateSchemaFor(doc, t.Elem())
		if err != nil {
			return nil, err
		}

		schema.Value = &openapi3.Schema{
			Type:  "array",
			Items: items,
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

		if doc.Components.Schemas == nil {
			doc.Components.Schemas = make(openapi3.Schemas)
		}

		// check if the structure already exists as component first
		_, found := doc.Components.Schemas[t.Name()]
		if !found {
			var sch *openapi3.Schema
			sch, err = generateStructureSchema(doc, t)
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

func generateStructureSchema(doc *openapi3.Swagger, t reflect.Type) (*openapi3.Schema, error) {
	ret := &openapi3.Schema{
		Type: "object",
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := parseJsonTag(f)

		if tag.Ignored {
			continue
		}

		fieldSchema, err := GenerateSchemaFor(doc, f.Type)
		if err != nil {
			return nil, err
		}

		ret.WithPropertyRef(tag.Name, fieldSchema)
	}

	// object type requires properties
	if ret.Properties != nil {
		ret.Type = "object"
	}

	return ret, nil
}

type jsonTag struct {
	OmitEmpty bool
	Name      string
	Ignored   bool
}

func parseJsonTag(f reflect.StructField) *jsonTag {
	ret := &jsonTag{
		Name: f.Name,
	}

	tag, found := f.Tag.Lookup("json")
	if found {
		values := strings.Split(tag, ",")
		for _, value := range values {
			switch value {
			case "-":
				ret.Ignored = true
			case "omitempty":
				ret.OmitEmpty = true
			default:
				ret.Name = value
			}
		}
	}

	return ret
}
