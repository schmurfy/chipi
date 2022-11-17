package builder

import (
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testQueryRequest struct {
	Path  struct{} `example:"/pet"`
	Query struct {
		Name string `chipi:"required"`
	}
}

func TestQueryParams(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Query", func() {
		var router *chi.Mux
		var b *Builder

		g.BeforeEach(func() {
			var err error
			router = chi.NewRouter()

			router.Get("/pet", emptyHandler)
			b, err = New(router, &openapi3.Info{})
			require.NoError(g, err)
		})

		g.Describe("from tags", func() {
			var op openapi3.Operation

			g.BeforeEach(func() {
				tt := reflect.TypeOf(testQueryRequest{})
				err := b.generateQueryParametersDoc(b.swagger, &op, tt, schema.Fields{})
				require.NoError(g, err)
			})

			g.Describe("Name", func() {
				var param *openapi3.Parameter
				g.BeforeEach(func() {
					param = op.Parameters.GetByInAndName("query", "name")
					require.NotNil(g, param)
					require.Equal(g, "name", param.Name)
				})

				g.It("should extract [required]", func() {
					assert.True(g, param.Required)
				})
			})
		})
	})
}
