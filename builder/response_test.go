package builder

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/response"
	"github.com/schmurfy/chipi/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Parent struct {
	Inline
	Field3 string
	Field4 int `json:"field4"`
}

type Inline struct {
	Field1 string
	Field2 int `json:"field2"`
}

func TestResponse(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("response documentation", func() {
		var b *Builder
		var op *openapi3.Operation
		var ctx context.Context

		g.BeforeEach(func() {
			var err error

			ctx = context.Background()

			router := chi.NewRouter()
			b, err = New(router, &openapi3.Info{})
			require.NoError(g, err)

			op = openapi3.NewOperation()
		})
		g.It("should return an error if structure does not implemenent ResponseEncoder", func() {
			req := struct {
				Response struct {
					Name string
				}
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.Error(g, err)
			assert.Contains(g, err.Error(), "must implement ResponseEncoder")
		})

		g.It("should add 204 response if there is no Response struct", func() {
			req := struct {
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			responseObj, found := op.Responses.Map()["204"]
			require.True(g, found)
			require.NotNil(g, responseObj)
		})

		g.It("should allow custom content-type", func() {
			req := struct {
				response.JsonEncoder
				Response struct {
					Name string
				} `content-type:"application/pdf"`
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			resp, found := op.Responses.Map()["200"]
			require.True(g, found)

			mediaType := resp.Value.Content.Get("application/json")
			require.Nil(g, mediaType)

			mediaType = resp.Value.Content.Get("application/pdf")
			require.NotNil(g, mediaType)

		})

		g.It("should handle json response", func() {
			req := struct {
				response.JsonEncoder
				Response struct {
					Name string
				}
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			resp, found := op.Responses.Map()["200"]
			require.True(g, found)

			mediaType := resp.Value.Content.Get("application/json")
			require.NotNil(g, mediaType)

			data, err := json.Marshal(mediaType.Schema)
			require.NoError(g, err)

			// check returned schema
			assert.JSONEq(g, `{
					"type": "object",
					"properties": {
						"Name": {
							"type": "string"
						}
					}
				}`, string(data))
		})

		g.It("should handle binary response", func() {
			req := struct {
				response.JsonEncoder
				Response []byte `description:"the requested file"`
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			resp, found := op.Responses.Map()["200"]
			require.True(g, found)

			assert.Equal(g, "the requested file", *resp.Value.Description)

			mediaType := resp.Value.Content.Get("application/json")
			require.Nil(g, mediaType)

			mediaType = resp.Value.Content.Get("application/octet-stream")
			require.NotNil(g, mediaType)
		})

		g.It("should handle array response", func() {
			req := struct {
				response.JsonEncoder
				Response []Inline
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			resp, found := op.Responses.Map()["200"]
			require.True(g, found)

			require.NotNil(g, *resp.Value.Content["application/json"].Schema.Value.Items)
		})

		g.It("should embed Inline struct", func() {
			req := struct {
				response.JsonEncoder
				Response Parent
			}{}

			err := b.generateResponseDoc(ctx, b.swagger, op, &req, reflect.TypeOf(req), shared.NewChipiCallbacks(nil))
			require.NoError(g, err)

			require.NotNil(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()])
			require.Len(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()].Value.Properties, 4)
			require.NotNil(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()].Value.Properties["Field1"])
			require.NotNil(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()].Value.Properties["field2"])
			require.NotNil(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()].Value.Properties["Field3"])
			require.NotNil(g, b.swagger.Components.Schemas[reflect.TypeOf(req.Response).String()].Value.Properties["field4"])
		})
	})
}
