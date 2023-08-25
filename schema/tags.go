package schema

import (
	"reflect"
	"strings"

	"github.com/schmurfy/chipi/shared"
)

type CastType string

const (
	CastTypeDatetime CastType = "datetime"
	CastTypeDuration CastType = "duration"
	CastTypeSchema   CastType = "schema"
)

var (
	SupportedTypeCasts = []CastType{CastTypeDatetime, CastTypeDuration, CastTypeSchema}
)

type jsonTag struct {
	Name string

	// from json or chipi tag
	OmitEmpty  *bool
	ReadOnly   *bool
	WriteOnly  *bool
	Nullable   *bool
	Ignored    *bool
	Deprecated *bool
	Required   *bool
	TypeCast   *CastType

	// self contained
	Explode     *bool
	Description *string
	Example     *string
	Style       *string
}

func ParseJsonTag(f reflect.StructField) *jsonTag {
	ret := &jsonTag{
		Name: f.Name,
	}

	if tag, found := f.Tag.Lookup("json"); found {
		values := strings.Split(tag, ",")
		for _, value := range values {
			switch value {
			case "-":
				ret.Ignored = boolPtr(true)
			case "omitempty":
				ret.OmitEmpty = boolPtr(true)
			default:
				ret.Name = value
			}
		}
	}

	if tag, found := f.Tag.Lookup("chipi"); found {
		values := strings.Split(tag, ",")
		for _, value := range values {
			switch value {
			case "ignore":
				ret.Ignored = boolPtr(true)
			case "readonly":
				ret.ReadOnly = boolPtr(true)
			case "writeonly":
				ret.WriteOnly = boolPtr(true)
			case "nullable":
				ret.Nullable = boolPtr(true)
			case "deprecated":
				ret.Deprecated = boolPtr(true)
			case "required":
				ret.Required = boolPtr(true)
			default:
				if strings.HasPrefix(value, "as:") {
					castType := CastType(strings.TrimPrefix(value, "as:"))
					if shared.Contains(SupportedTypeCasts, castType) {
						ret.TypeCast = &castType
					}
				}
			}
		}
	}

	if val, found := f.Tag.Lookup("description"); found {
		ret.Description = stringPtr(val)
	}

	if val, found := f.Tag.Lookup("example"); found {
		ret.Example = stringPtr(val)
	}

	if val, found := f.Tag.Lookup("style"); found {
		ret.Style = stringPtr(val)
	}

	if val, found := f.Tag.Lookup("explode"); found {
		b := (val == "true")
		ret.Explode = &b
	}

	return ret
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
