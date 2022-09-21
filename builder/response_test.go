package builder

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("response documentation", func() {
		var b *Builder
		var op *openapi3.Operation

		g.BeforeEach(func() {
			var err error

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

			err := b.generateResponseDoc(b.swagger, op, &req, reflect.TypeOf(req))
			require.Error(g, err)
			assert.Contains(g, err.Error(), "must implement ResponseEncoder")
		})

		g.It("should allow custom content-type", func() {
			req := struct {
				response.JsonEncoder
				Response struct {
					Name string
				} `content-type:"application/pdf"`
			}{}

			err := b.generateResponseDoc(b.swagger, op, &req, reflect.TypeOf(req))
			require.NoError(g, err)

			resp, found := op.Responses["200"]
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

			err := b.generateResponseDoc(b.swagger, op, &req, reflect.TypeOf(req))
			require.NoError(g, err)

			resp, found := op.Responses["200"]
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

			err := b.generateResponseDoc(b.swagger, op, &req, reflect.TypeOf(req))
			require.NoError(g, err)

			resp, found := op.Responses["200"]
			require.True(g, found)

			assert.Equal(g, "the requested file", *resp.Value.Description)

			mediaType := resp.Value.Content.Get("application/json")
			require.Nil(g, mediaType)
		})

	})
}
