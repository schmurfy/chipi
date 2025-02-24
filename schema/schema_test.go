package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/schmurfy/chipi/internal/testdata/monster"
	"github.com/schmurfy/chipi/internal/testdata/pet"
	"github.com/schmurfy/chipi/shared"
)

type RecursiveGroup struct {
	Name  string
	Users []*RecursiveUser
}

type RecursiveUser struct {
	Name  string
	Group *RecursiveGroup
}

type Generic[T any] struct {
	Name  string
	Value T
}

type Embedded struct {
}

func checkGeneratedType(g *goblin.G, ctx context.Context, schemaPtr **Schema, docPtr **openapi3.T, value interface{}, expected string) {
	g.It(fmt.Sprintf("should generate inline type for %T", value), func() {
		s := *schemaPtr
		doc := *docPtr

		require.NotNil(g, doc)
		require.NotNil(g, s)

		typ := reflect.TypeOf(value)
		schema, err := s.GenerateSchemaFor(ctx, doc, typ)
		require.NoError(g, err)

		data, err := json.Marshal(schema)
		require.NoError(g, err)

		assert.JSONEq(g, expected, string(data))
	})

}

type TestFilter struct {
	AllowedFields []string
}

var _ shared.FilterFieldInterface = &TestFilter{}
var _ shared.SchemaResolverInterface = &TestFilter{}

func (f *TestFilter) FilterField(ctx context.Context, fieldInfo shared.AttributeInfo) (bool, error) {
	for _, path := range f.AllowedFields {
		if path == fieldInfo.QueryPath() {
			return false, nil
		}
	}
	return true, nil
}

func (f *TestFilter) SchemaResolver(fieldInfo shared.AttributeInfo, castName string, _ reflect.Type, _ shared.GenerateSchemaCallbackType) (*openapi3.Schema, bool) {
	switch castName {
	case "datetime":
		return openapi3.NewDateTimeSchema(), false
	case "duration":
		return openapi3.NewInt64Schema(), false
	default:
		return nil, false
	}
}

type TestEnumResolver struct {
}

var _ shared.EnumResolverInterface = &TestEnumResolver{}
var _ shared.SchemaResolverInterface = &TestEnumResolver{}

func (e *TestEnumResolver) EnumResolver(t reflect.Type) (bool, shared.Enum) {
	if t.Name() == "UserGender" {
		return true, []shared.EnumEntry{
			{Title: "NOT_SET", Value: 0},
			{Title: "MALE", Value: 1},
			{Title: "FEMALE", Value: 2},
		}
	}
	return false, nil
}

func (e *TestEnumResolver) SchemaResolver(fieldInfo shared.AttributeInfo, castName string, _ reflect.Type, _ shared.GenerateSchemaCallbackType) (*openapi3.Schema, bool) {
	switch castName {
	case "datetime":
		return openapi3.NewDateTimeSchema(), false
	case "duration":
		return openapi3.NewInt64Schema(), false
	default:
		return nil, false
	}
}

func TestSchema(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("schema", func() {
		var doc *openapi3.T
		var s *Schema
		var ctx context.Context

		g.BeforeEach(func() {
			var err error

			ctx = context.Background()

			doc = &openapi3.T{}
			s, err = New()
			require.NoError(g, err)
		})

		g.Describe("basic types", func() {
			tests := map[interface{}]string{
				"":       `{"type": "string"}`,
				false:    `{"type": "boolean"}`,
				3.14:     `{"type": "number", "format": "double"}`,
				int32(2): `{"type": "integer", "format": "int32"}`,
				int(42):  `{"type": "integer", "format": "int64"}`,
			}

			for value, expected := range tests {
				checkGeneratedType(g, ctx, &s, &doc, value, expected)
			}
		})

		g.Describe("complex types", func() {
			tests := []struct {
				Name     string
				Value    interface{}
				Expected string
			}{
				{Name: "[]string", Value: []string{}, Expected: `{
					"type": "array", "items": {
						"type": "string"
					}
				}`},

				{Name: "[]bool", Value: []bool{}, Expected: `{
					"type": "array", "items": {
						"type": "boolean"
					}
				}`},

				{Name: "[]int32", Value: []int32{4, 5}, Expected: `{
					"type": "array", "items": {
						"type": "integer",
						"format": "int32"
					}
				}`},
			}

			for _, tt := range tests {
				checkGeneratedType(g, ctx, &s, &doc, tt.Value, tt.Expected)
			}
		})

		g.Describe("different packages", func() {
			g.It("should generate correct reference path", func() {
				typ1 := reflect.TypeOf(monster.QueryResponse{})
				path := schemaReference(typ1)
				assert.Equal(g, "#/components/schemas/monster.QueryResponse", path)
			})

			g.It("should generate correct types for same structure name", func() {
				typ1 := reflect.TypeOf(&monster.QueryResponse{})
				schema1, err := s.GenerateSchemaFor(ctx, doc, typ1)
				require.NoError(g, err)

				typ2 := reflect.TypeOf(&pet.QueryResponse{})
				schema2, err := s.GenerateSchemaFor(ctx, doc, typ2)
				require.NoError(g, err)

				assert.NotEqual(g, schema1.Ref, schema2.Ref)
			})
		})

		g.Describe("structures", func() {
			type UserGender int
			type WrappedTime struct {
				time.Time
			}
			type WrappedDuration struct {
				Duration int64
			}
			type User struct {
				Name     string `json:"name,omitempty"`
				Age      int
				Ignored  bool `json:"-"`
				Sex      UserGender
				Time     WrappedTime     `chipi:"as:datetime"`
				Duration WrappedDuration `chipi:"as:duration"`
			}

			type Group struct {
				Name  string
				Users []User
			}

			g.It("should filter direct fields", func() {
				filter := &TestFilter{AllowedFields: []string{
					"user",
					"user.age",
				}}

				typ := reflect.TypeOf(&User{})
				schema, err := s.GenerateFilteredSchemaFor(ctx, doc, typ, shared.NewChipiCallbacks(filter))
				require.NoError(g, err)

				_, err = json.Marshal(schema)
				require.NoError(g, err)

				userSchema, found := doc.Components.Schemas[typeName(typ.Elem())]
				require.True(g, found)

				data, err := json.Marshal(userSchema)
				require.NoError(g, err)

				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"Age": {
							"type": "integer",
							"format": "int64"
						}
					}
				}`, string(data))
			})

			g.It("should use oneof as refs for enums and cast fields with as:", func() {

				typ := reflect.TypeOf(&User{})
				schema, err := s.GenerateFilteredSchemaFor(ctx, doc, typ, shared.NewChipiCallbacks(&TestEnumResolver{}))
				require.NoError(g, err)

				_, err = json.Marshal(schema)
				require.NoError(g, err)

				userSchema, found := doc.Components.Schemas[typeName(typ.Elem())]
				require.True(g, found)

				data, err := json.Marshal(userSchema)
				require.NoError(g, err)

				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"name": {
							"type": "string"
						},
						"Age": {
							"type": "integer",
							"format": "int64"
						},
						"Time": {"format":"date-time", "type":"string"},
						"Duration": {"format":"int64", "type":"integer"},
						"Sex": {"$ref":"#/components/schemas/schema.UserGender"}
					}
				}`, string(data))

				ref, err := json.Marshal(doc.Components.Schemas["schema.UserGender"])
				require.NoError(g, err)
				assert.JSONEq(g, `{
					"type": "integer",
					"format": "int64",
					"oneOf": [
						{ "type": "const", "format": "int64", "const": 0, "title": "NOT_SET" },
						{ "type": "const", "format": "int64", "const": 1, "title": "MALE" },
						{ "type": "const", "format": "int64", "const": 2, "title": "FEMALE" }
					]
				}`, string(ref))
			})

			g.It("should generate referenced type for user", func() {
				typ := reflect.TypeOf(&User{})
				schema, err := s.GenerateFilteredSchemaFor(ctx, doc, typ, shared.NewChipiCallbacks(&TestEnumResolver{}))
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"$ref": "#/components/schemas/schema.User"
				}`, string(data))

				// check that the User schema was added as component
				userSchema, found := doc.Components.Schemas[typeName(typ.Elem())]
				require.True(g, found)

				data, err = json.Marshal(userSchema)
				require.NoError(g, err)

				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"name": {
							"type": "string"
						},
						"Age": {
							"type": "integer",
							"format": "int64"
						},
						"Sex": {"$ref":"#/components/schemas/schema.UserGender"},
						"Time": {"format":"date-time", "type":"string"},
						"Duration": {"format":"int64", "type":"integer"}
					}
				}`, string(data))
			})

			g.It("should handle recursive structures", func() {

				g.Timeout(5 * time.Second)
				typ := reflect.TypeOf(&RecursiveUser{})
				_, err := s.GenerateSchemaFor(ctx, doc, typ)
				require.NoError(g, err)
			})

			// type UploadResumeRequest struct {
			// 	Path struct {
			// 		Name string `example:"john"`
			// 	} `example:"/user/john"`
			// 	Query struct{}

			// 	Body struct {
			// 		File1 []byte
			// 		File2 []byte
			// 	} `content-type:"multipart/form-data"`
			// }

			g.It("should inline anonymous structures", func() {
				st := struct {
					Cool bool
				}{}

				typ := reflect.TypeOf(&st)
				schema, err := s.GenerateSchemaFor(ctx, doc, typ)
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"Cool": {
							"type": "boolean"
						}
					}
				}`, string(data))
			})

			g.It("should generate referenced type for Group with link to User", func() {
				typ := reflect.TypeOf(&Group{})
				schema, err := s.GenerateSchemaFor(ctx, doc, typ)
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"$ref": "#/components/schemas/schema.Group"
				}`, string(data))

				// check that the User schema was added as component

				_, found := doc.Components.Schemas[typeName(reflect.TypeOf(&User{}))]
				require.True(g, found)

				userSchema, found := doc.Components.Schemas[typeName(typ)]
				require.True(g, found)

				data, err = json.Marshal(userSchema)
				require.NoError(g, err)

				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"Name": {
							"type": "string"
						},
						"Users": {
							"type": "array",
							"items": {
								"$ref": "#/components/schemas/schema.User"
							}
						}
					}
				}`, string(data))
			})

			g.It("should generate valid name for generic structures", func() {
				st := Generic[Embedded]{
					Name: "john",
				}
				typ := reflect.TypeOf(&st)
				schema, err := s.GenerateSchemaFor(ctx, doc, typ)
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)
				// check returned schema
				assert.JSONEq(g, `{
					"$ref": "#/components/schemas/schema.Generic..schema.Embedded"
				}`, string(data))

				userSchema, found := doc.Components.Schemas[typeName(typ.Elem())]
				require.True(g, found)

				data, err = json.Marshal(userSchema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"Name": {
							"type": "string"
						},
						"Value": {
							"$ref": "#/components/schemas/schema.Embedded"
						}
					}
				}`, string(data))
			})

			checkGeneratedType(g, ctx, &s, &doc, time.Time{}, `{
				"type": "string",
				"format": "date-time"
			}`)
		})

	})
}
