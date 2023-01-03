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

func (f *TestFilter) FilterRoute(ctx context.Context, method string, pattern string) (bool, error) {
	return false, nil
}

func (f *TestFilter) FilterField(ctx context.Context, fieldInfo shared.AttributeInfo) (bool, error) {
	for _, path := range f.AllowedFields {
		if path == fieldInfo.QueryPath() {
			return false, nil
		}
	}
	return true, nil
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
				path := structReference(typ1)
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
			type User struct {
				Name    string `json:"name,omitempty"`
				Age     int
				Ignored bool `json:"-"`
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
				schema, err := s.GenerateFilteredSchemaFor(ctx, doc, typ, filter)
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

			g.It("should generate referenced type for user", func() {
				typ := reflect.TypeOf(&User{})
				schema, err := s.GenerateSchemaFor(ctx, doc, typ)
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
						}
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

			checkGeneratedType(g, ctx, &s, &doc, time.Time{}, `{
				"type": "string",
				"format": "date-time"
			}`)

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
		})

	})
}
