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

		})

	})
}
