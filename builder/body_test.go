package builder

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopHandler struct{}

func (e *noopHandler) Handle(context.Context, http.ResponseWriter) error {
	return nil
}

type bodyTestWithoutDecoderRequest struct {
	noopHandler

	Path struct {
	} `example:"/pet"`

	Body struct {
		Name string
	}
}

type bodyTestWithDecoderRequest struct {
	noopHandler

	Path struct {
	} `example:"/pet"`

	request.JsonBodyDecoder
	Body struct {
		Name  string
		Email string
		Age   int
	}
}

func TestBodyGenerator(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("body generator", func() {
		var b *Builder
		var op openapi3.Operation
		var ctx context.Context

		g.BeforeEach(func() {
			var err error
			router := chi.NewRouter()

			ctx = context.Background()

			router.Post("/pet/{Id}/{Name}", emptyHandler)
			b, err = New(router, &openapi3.Info{})
			require.NoError(g, err)
		})

		g.It("should return an error if structure does not implements BodyDecoder", func() {
			req := bodyTestWithoutDecoderRequest{}
			err := b.generateBodyDoc(ctx, b.swagger, &op, &req, reflect.TypeOf(req), nil)
			require.Error(g, err)

			assert.Contains(g, err.Error(), "must implement BodyDecoder")
		})

		g.It("should return nil if structure implements BodyDecoder", func() {
			req := bodyTestWithDecoderRequest{}
			err := b.generateBodyDoc(ctx, b.swagger, &op, &req, reflect.TypeOf(req), nil)
			require.NoError(g, err)
		})

	})
}
