package schema

import (
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
)

type RecursiveGroup struct {
	Name  string
	Users []*RecursiveUser
}

type RecursiveUser struct {
	Name  string
	Group *RecursiveGroup
}

func checkGeneratedType(g *goblin.G, schemaPtr **Schema, docPtr **openapi3.T, value interface{}, expected string) {
	g.It(fmt.Sprintf("should generate inline type for %T", value), func() {
		s := *schemaPtr
		doc := *docPtr

		require.NotNil(g, doc)
		require.NotNil(g, s)

		typ := reflect.TypeOf(value)
		schema, err := s.GenerateSchemaFor(doc, typ, []string{})
		require.NoError(g, err)

		data, err := json.Marshal(schema)
		require.NoError(g, err)

		assert.JSONEq(g, expected, string(data))
	})

}

func TestSchema(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("schema", func() {
		var doc *openapi3.T
		var s *Schema

		g.BeforeEach(func() {
			var err error

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
				checkGeneratedType(g, &s, &doc, value, expected)
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
				checkGeneratedType(g, &s, &doc, tt.Value, tt.Expected)
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
				schema1, err := s.GenerateSchemaFor(doc, typ1, []string{})
				require.NoError(g, err)

				typ2 := reflect.TypeOf(&pet.QueryResponse{})
				schema2, err := s.GenerateSchemaFor(doc, typ2, []string{})
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

			g.It("should generate referenced type for user", func() {
				typ := reflect.TypeOf(&User{})
				schema, err := s.GenerateSchemaFor(doc, typ, []string{})
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
				_, err := s.GenerateSchemaFor(doc, typ, []string{})
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
				schema, err := s.GenerateSchemaFor(doc, typ, []string{})
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
				schema, err := s.GenerateSchemaFor(doc, typ, []string{})
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"$ref": "#/components/schemas/schema.Group"
				}`, string(data))

				// check that the User schema was added as component

				userSchema, found := doc.Components.Schemas[typeName(reflect.TypeOf(&User{}))]
				require.True(g, found)

				userSchema, found = doc.Components.Schemas[typeName(typ)]
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

			checkGeneratedType(g, &s, &doc, time.Time{}, `{
				"type": "string",
				"format": "date-time"
			}`)
		})

	})
}
