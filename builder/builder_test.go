package builder

import (
	"context"
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/response"
	"github.com/schmurfy/chipi/shared"
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

type TestRoute struct {
	Method  string
	Pattern string
}

type TestFilter struct {
	AllowedRoutes []TestRoute
	AllowedFields []string
}

func (f *TestFilter) FilterRoute(ctx context.Context, method string, pattern string) (bool, error) {
	for _, rr := range f.AllowedRoutes {
		if (rr.Method == method) && (rr.Pattern == pattern) {
			return false, nil
		}
	}

	return true, nil
}

func (f *TestFilter) FilterField(ctx context.Context, fieldInfo shared.AttributeInfo) (bool, error) {
	for _, path := range f.AllowedFields {
		if path == fieldInfo.QueryPath() {
			return false, nil
		}
	}
	return true, nil
}

func TestBuilder(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Builder", func() {

		g.Describe("nested routers", func() {
			var b *Builder
			var router *chi.Mux
			var ctx context.Context

			g.BeforeEach(func() {
				var err error
				router = chi.NewRouter()

				ctx = context.Background()

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
					json, err := b.GenerateJson(ctx, shared.NewChipiCallbacks(nil))
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.NotNil(g, swagger.Paths.Map()[routePath])
				})
				g.It("should filter routes", func() {
					filter := TestFilter{AllowedRoutes: []TestRoute{
						{Method: "POST", Pattern: "other/route"},
					}}

					json, err := b.GenerateJson(ctx, shared.NewChipiCallbacks(&filter))
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.Nil(g, swagger.Paths.Map()[routePath])
				})

				g.It("should authorize routes", func() {
					filter := TestFilter{AllowedRoutes: []TestRoute{
						{Method: "POST", Pattern: routePath},
					}}

					json, err := b.GenerateJson(ctx, shared.NewChipiCallbacks(&filter))
					require.Nil(g, err)

					swagger := convertToSwagger(g, json)

					require.NotNil(g, swagger.Paths.Map()[routePath])
				})

			})

		})

	})
}
