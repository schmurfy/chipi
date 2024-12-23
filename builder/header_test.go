package builder

import (
	"context"
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHeaderRequest struct {
	Path struct {
		Id int
	} `example:"/pet/43/Fido"`

	Header struct {
		XClientId string `name:"X-ClientId"`
	}
}

func TestHeader(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Header", func() {
		var router *chi.Mux
		var b *Builder
		var ctx context.Context

		g.Describe("single router", func() {

			g.BeforeEach(func() {
				var err error
				router = chi.NewRouter()

				ctx = context.Background()

				router.Post("/pet/{Id}/{Name}", emptyHandler)
				b, err = New(router, &openapi3.Info{})
				require.NoError(g, err)
			})

			g.Describe("from tags", func() {
				var op openapi3.Operation

				g.BeforeEach(func() {
					tt := reflect.TypeOf(testHeaderRequest{})
					// routeContext := chi.NewRouteContext()
					// require.True(g, router.Match(routeContext, "POST", "/pet/43/Fido"))
					err := b.generateHeadersDoc(ctx, b.swagger, &op, tt)
					require.NoError(g, err)
				})

				g.Describe("X-ClientId", func() {
					var param *openapi3.Parameter
					g.BeforeEach(func() {
						param = op.Parameters.GetByInAndName(openapi3.ParameterInHeader, "X-ClientId")
						require.NotNil(g, param)
					})

					g.It("should override field name with [name]", func() {
						assert.Equal(g, "X-ClientId", param.Name)
					})

				})

			})

		})
	})

}
