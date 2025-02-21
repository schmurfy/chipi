package schema

import (
	"reflect"
	"strings"
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
	CastName   *string

	// self contained
	Explode     *bool
	Description *string
	Example     *string
	Style       *string
}

func (t *jsonTag) GetOmitEmpty() bool {
	if t.OmitEmpty == nil {
		return false
	}
	return *t.OmitEmpty
}
func (t *jsonTag) GetReadOnly() bool {
	if t.ReadOnly == nil {
		return false
	}
	return *t.ReadOnly
}
func (t *jsonTag) GetWriteOnly() bool {
	if t.WriteOnly == nil {
		return false
	}
	return *t.WriteOnly
}
func (t *jsonTag) GetExplode() bool {
	if t.Explode == nil {
		return false
	}
	return *t.Explode
}
func (t *jsonTag) GetNullable() bool {
	if t.Nullable == nil {
		return false
	}
	return *t.Nullable
}
func (t *jsonTag) GetIgnored() bool {
	if t.Ignored == nil {
		return false
	}
	return *t.Ignored
}
func (t *jsonTag) GetDeprecated() bool {
	if t.Deprecated == nil {
		return false
	}
	return *t.Deprecated
}
func (t *jsonTag) GetRequired() bool {
	if t.Required == nil {
		return false
	}
	return *t.Required
}
func (t *jsonTag) GetCastName() string {
	if t.CastName == nil {
		return ""
	}
	return *t.CastName
}
func (t *jsonTag) GetDescription() string {
	if t.Description == nil {
		return ""
	}
	return *t.Description
}
func (t *jsonTag) GetExample() string {
	if t.Example == nil {
		return ""
	}
	return *t.Example
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
					castName := strings.TrimPrefix(value, "as:")
					ret.CastName = &castName
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
