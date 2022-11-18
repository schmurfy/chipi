package builder

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/request"
	"github.com/schmurfy/chipi/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopHandler struct{}

type GolangTypes struct {
	ID     string
	Int    int
	IntPtr *int

	Int8    int8
	Int8Ptr *int8

	Int16 int16
	Int32 int32
	Int64 int64

	Uint   uint
	Uint32 uint32
	Uint64 uint64

	Float32 float32
	Float64 float64

	ArrString    []string
	ArrStringPtr []*string

	Str    string
	StrPtr *string

	Bool    bool
	BoolPtr *bool

	Nested    Nested
	NestedPtr *Nested

	ArrNested    []Nested
	ArrNestedPtr []*Nested

	ArrNested2 []Nested2

	Time *time.Time

	//External external_package_test.External
}

type Nested struct {
	Type    string
	Nnested NestedNested
}

type NestedNested struct {
	Bool bool
}

type Nested2 struct {
	String string
}

func (e *noopHandler) Handle(context.Context, http.ResponseWriter) error {
	return nil
}

type bodyTestWithStructBody struct {
	noopHandler

	Path struct {
	} `example:"/pet"`
	request.JsonBodyDecoder

	Body GolangTypes
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
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), schema.Fields{})
			require.Error(g, err)

			assert.Contains(g, err.Error(), "must implement BodyDecoder")
		})

		g.It("should return nil if structure implements BodyDecoder", func() {
			req := bodyTestWithDecoderRequest{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), schema.Fields{})
			require.NoError(g, err)
		})

		g.It("should protect a field", func() {
			req := bodyTestWithStructBody{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), schema.Fields{
				Protected: []string{"builder.golangtypes.id"},
			})
			//If builder.anystruct.id is passed to generateBodyDoc, it means the user doesn't have the permissions
			require.NoError(g, err)
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["ID"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Int"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["IntPtr"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Str"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["StrPtr"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Bool"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["BoolPtr"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Nested"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["NestedPtr"])
			require.NotNil(g, b.swagger.Components.Schemas["builder.Nested"])
		})

		g.It("should whitelist a field", func() {
			req := bodyTestWithStructBody{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), schema.Fields{
				Whitelist: []string{"builder.golangtypes.id"},
			})
			require.NoError(g, err)
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["ID"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Int"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Int"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["IntPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Str"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["StrPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Bool"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["BoolPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Nested"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["NestedPtr"])
		})

		g.It("should still whitelists fields", func() {
			req := bodyTestWithStructBody{}
			err := b.generateBodyDoc(b.swagger, &op, &req, reflect.TypeOf(req), schema.Fields{
				Whitelist: []string{"builder.golangtypes.id"},
				Protected: []string{"builder.golangtypes.int"},
			})
			require.NoError(g, err)
			require.NotNil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["ID"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Int"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Int"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["IntPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Str"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["StrPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Bool"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["BoolPtr"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["Nested"])
			require.Nil(g, b.swagger.Components.Schemas["builder.GolangTypes"].Value.Properties["NestedPtr"])
		})
	})
}
