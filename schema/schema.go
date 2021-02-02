package schema

import (
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	_timeType = reflect.TypeOf(time.Time{})
)

func GenerateSchemaFor(doc *openapi3.Swagger, t reflect.Type) (schema *openapi3.Schema, err error) {
	// Get TypeInfo
	// typeInfo := jsoninfo.GetTypeInfo(t)

	schema = &openapi3.Schema{}

	// test pointed value for pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {

	// basic types
	case reflect.String:
		schema.Type = "string"

	case reflect.Bool:
		schema.Type = "boolean"

	case reflect.Int8, reflect.Int16, reflect.Int32:
		schema.Type = "integer"
		schema.Format = "int32"

	case reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
		schema.Format = "int64"

	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
		schema.Format = "double"

	// complex types
	case reflect.Slice:
		schema.Type = "array"
		items, err := GenerateSchemaFor(doc, t.Elem())
		if err != nil {
			return nil, err
		}

		schema.Items = &openapi3.SchemaRef{Value: items}

	case reflect.Map:
		panic("todo")

	case reflect.Struct:
		if t == _timeType {
			schema.Type = "string"
			schema.Format = "date-time"
			return
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

			schema.WithPropertyRef(tag.Name, openapi3.NewSchemaRef("", fieldSchema))
		}

		// object type requires properties
		if schema.Properties != nil {
			schema.Type = "object"
		}
	}

	return
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
