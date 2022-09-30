package builder

import (
	"context"
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/response"
	"github.com/stretchr/testify/require"
)

type builderTestPathRequest struct {
	response.ErrorEncoder

	Path struct {
		Id int
	} `example:"/pets/43"`
}

func (r *builderTestPathRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	return nil
}

func convertToSwagger(g *goblin.G, data []byte) *openapi3.T {
	swagger := &openapi3.T{
		OpenAPI: "3.1.0",
	}
	err := swagger.UnmarshalJSON(data)
	require.Nil(g, err)
	return swagger
}
func TestBuilder(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Builder", func() {

		g.Describe("nested routers", func() {
			var b *Builder
			var router *chi.Mux

			g.BeforeEach(func() {
				var err error
				router = chi.NewRouter()

				b, err = New(router, &openapi3.Info{})
				require.NoError(g, err)
			})

			g.It("should detect direct path", func() {
				err := b.Post(router, "/pets/{Id}", &builderTestPathRequest{})
				require.NoError(g, err)
			})

			g.It("should detect nested path", func() {
				petsRoute := chi.NewRouter()
				router.Mount("/pets", petsRoute)
				petsRoute.Group(func(r chi.Router) {
					err := b.Post(r, "/{Id}", &builderTestPathRequest{})
					require.NoError(g, err)
				})

			})

			g.Describe("test filter routes", func() {

				routePath := "/pets/{Id}"
				g.BeforeEach(func() {
					err := b.Post(router, routePath, &builderTestPathRequest{})
					require.NoError(g, err)
				})

				g.It("should not filter routes", func() {
					json, err := b.GenerateJson(false, []string{""}, []string{""})
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.NotNil(g, swagger.Paths[routePath])
				})
				g.It("should filter routes", func() {
					json, err := b.GenerateJson(true, []string{"POST other/route"}, []string{""})
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.Nil(g, swagger.Paths[routePath])
				})

				g.It("should authorize routes", func() {
					json, err := b.GenerateJson(true, []string{"POST " + routePath}, []string{""})
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.NotNil(g, swagger.Paths[routePath])
				})

			})

		})

	})
}
