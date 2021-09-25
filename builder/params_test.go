package builder

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPathRequest struct {
	Path struct {
		Id   int
		Name string `example:"Ralph" description:"some text" deprecated:"true" style:"tarzan" explode:"true"`
	} `example:"/pet/43/Fido"`
}

func emptyHandler(w http.ResponseWriter, r *http.Request) {

}

func TestParams(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Params", func() {
		var router *chi.Mux
		var b *Builder

		g.BeforeEach(func() {
			router = chi.NewRouter()

			router.Post("/pet/{Id}/{Name}", emptyHandler)
			b = &Builder{} // is it really required ?

		})

		g.Describe("from tags", func() {
			var op openapi3.Operation

			g.BeforeEach(func() {
				tt := reflect.TypeOf(testPathRequest{})
				err := b.generateParametersDoc(router, &op, tt, "POST")
				require.NoError(g, err)
			})

			g.Describe("Id (control group for defaults)", func() {
				var param *openapi3.Parameter
				g.BeforeEach(func() {
					param = op.Parameters.GetByInAndName("path", "Id")
					require.NotNil(g, param)
					require.Equal(g, "Id", param.Name)
				})

				g.It("should extract [example]", func() {
					assert.Nil(g, param.Example)
				})

				g.It("should extract [description]", func() {
					assert.Equal(g, "", param.Description)
				})

				g.It("should extract [deprecated]", func() {
					assert.Equal(g, false, param.Deprecated)
				})

				g.It("should extract [style]", func() {
					assert.Equal(g, "", param.Style)
				})

				g.It("should extract [explode]", func() {
					assert.Nil(g, param.Explode)
				})

			})

			g.Describe("Name", func() {
				var param *openapi3.Parameter
				g.BeforeEach(func() {
					param = op.Parameters.GetByInAndName("path", "Name")
					require.NotNil(g, param)
					require.Equal(g, "Name", param.Name)
				})

				g.It("should extract [example]", func() {
					assert.Equal(g, "Ralph", param.Example)
				})

				g.It("should extract [description]", func() {
					assert.Equal(g, "some text", param.Description)
				})

				g.It("should extract [deprecated]", func() {
					assert.Equal(g, true, param.Deprecated)
				})

				g.It("should extract [style]", func() {
					assert.Equal(g, "tarzan", param.Style)
				})

				g.It("should extract [explode]", func() {
					require.NotNil(g, param.Explode)
					assert.Equal(g, true, *param.Explode)
				})

			})

		})

	})
}
