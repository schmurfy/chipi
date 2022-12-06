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

type testQueryRequest struct {
	Path  struct{} `example:"/pet"`
	Query struct {
		Name                  string `chipi:"required"`
		NoJsonTag             string
		SnakeCaseWithJsonTag  string `json:"overrided_name_with_tag"`
		PascalCaseWithJsonTag string `json:"PascalCaseWithJsonTag"`
		CamelCaseWithJsonTag  string `json:"camelCaseWithJsonTag"`
		PascalCaseWithNameTag string `name:"PascalCaseWithNameTag"`
	}
}

func TestQueryParams(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Query", func() {
		var router *chi.Mux
		var b *Builder
		var ctx context.Context

		g.BeforeEach(func() {
			var err error
			router = chi.NewRouter()

			ctx = context.Background()

			router.Get("/pet", emptyHandler)
			b, err = New(router, &openapi3.Info{})
			require.NoError(g, err)
		})

		g.Describe("from tags", func() {
			var op openapi3.Operation

			g.BeforeEach(func() {
				tt := reflect.TypeOf(testQueryRequest{})
				err := b.generateQueryParametersDoc(ctx, b.swagger, &op, tt)
				require.NoError(g, err)
			})

			g.Describe("Name", func() {
				var param *openapi3.Parameter
				var paramPascalCaseNoJsonTag *openapi3.Parameter
				var paramSnakeCaseWithJsonTag *openapi3.Parameter
				var paramPascalCaseWithJsonTag *openapi3.Parameter
				var paramCamelCaseWithJsonTag *openapi3.Parameter
				var paramPascalCaseWithNameTag *openapi3.Parameter

				g.It("test parsing json tag", func() {
					param = op.Parameters.GetByInAndName("query", "name")
					require.NotNil(g, param)
					require.Equal(g, "name", param.Name)

					paramPascalCaseNoJsonTag = op.Parameters.GetByInAndName("query", "no_json_tag")
					require.NotNil(g, paramPascalCaseNoJsonTag)
					require.Equal(g, "no_json_tag", paramPascalCaseNoJsonTag.Name)

					paramSnakeCaseWithJsonTag = op.Parameters.GetByInAndName("query", "overrided_name_with_tag")
					require.NotNil(g, paramSnakeCaseWithJsonTag)
					require.Equal(g, "overrided_name_with_tag", paramSnakeCaseWithJsonTag.Name)

					paramPascalCaseWithJsonTag = op.Parameters.GetByInAndName("query", "pascal_case_with_json_tag")
					require.NotNil(g, paramPascalCaseWithJsonTag)
					require.Equal(g, "pascal_case_with_json_tag", paramPascalCaseWithJsonTag.Name)

					paramCamelCaseWithJsonTag = op.Parameters.GetByInAndName("query", "camelCaseWithJsonTag")
					require.NotNil(g, paramCamelCaseWithJsonTag)
					require.Equal(g, "camelCaseWithJsonTag", paramCamelCaseWithJsonTag.Name)

					paramPascalCaseWithNameTag = op.Parameters.GetByInAndName("query", "pascal_case_with_name_tag")
					require.NotNil(g, paramPascalCaseWithNameTag)
					require.Equal(g, "pascal_case_with_name_tag", paramPascalCaseWithNameTag.Name)
				})

				g.It("should extract [required]", func() {
					assert.True(g, param.Required)
				})
			})
		})
	})
}
