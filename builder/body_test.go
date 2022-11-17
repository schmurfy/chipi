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

type AnyStruct struct {
	Name  string
	Name2 string
}

func (e *noopHandler) Handle(context.Context, http.ResponseWriter) error {
	return nil
}

type bodyTestWithStructBody struct {
	noopHandler

	Path struct {
	} `example:"/pet"`
	request.JsonBodyDecoder

	Body AnyStruct
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
		Name string
	}
}

func TestBodyGenerator(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("body generator", func() {
		var b *Builder
		var op openapi3.Operation

		g.BeforeEach(func() {
			var err error
			router := chi.NewRouter()

			router.Post("/pet/{Id}/{Name}", emptyHandler)
			b, err = New(router, &openapi3.Info{})
			require.NoError(g, err)
		})

		g.It("should return an error if structure does not implements BodyDecoder", func() {
			req := bodyTestWithoutDecoderRequest{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), []string{})
			require.Error(g, err)

			assert.Contains(g, err.Error(), "must implement BodyDecoder")
		})

		g.It("should return nil if structure implements BodyDecoder", func() {
			req := bodyTestWithDecoderRequest{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), []string{})
			require.NoError(g, err)
		})

		g.It("should ?? blacklist", func() {
			req := bodyTestWithStructBody{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), []string{"builder.anystruct.name"})
			require.NoError(g, err)
			require.Nil(g, b.swagger.Components.Schemas["builder.AnyStruct"].Value.Properties["Name"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.AnyStruct"].Value.Properties["Name2"])
		})
	})
}
