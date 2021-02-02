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
)

func checkGeneratedType(g *goblin.G, doc *openapi3.Swagger, value interface{}, expected string) {
	g.It(fmt.Sprintf("should generate inline type for %T", value), func() {
		typ := reflect.TypeOf(value)
		schema, err := GenerateSchemaFor(doc, typ)
		require.NoError(g, err)

		data, err := json.Marshal(schema)
		require.NoError(g, err)

		assert.JSONEq(g, expected, string(data))
	})

}

func TestSchema(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("schema", func() {
		var doc *openapi3.Swagger

		g.BeforeEach(func() {
			doc = &openapi3.Swagger{}
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
				checkGeneratedType(g, doc, value, expected)
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
				checkGeneratedType(g, doc, tt.Value, tt.Expected)
			}
		})

		g.Describe("structures", func() {
			type User struct {
				Name    string `json:"name,omitempty"`
				Age     int
				Ignored bool `json:"-"`
			}

			g.It("should generate referenced type for user", func() {
				typ := reflect.TypeOf(&User{})
				schema, err := GenerateSchemaFor(doc, typ)
				require.NoError(g, err)

				data, err := json.Marshal(schema)
				require.NoError(g, err)

				// check returned schema
				assert.JSONEq(g, `{
					"$ref": "#/components/schemas/User"
				}`, string(data))

				// check that the User schema was added as component

				userSchema, found := doc.Components.Schemas["User"]
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

			checkGeneratedType(g, doc, time.Time{}, `{
				"type": "string",
				"format": "date-time"
			}`)
		})

	})
}
