package builder

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/shared"
	"github.com/stretchr/testify/require"
)

type MultiMimeRequest struct {
	Path struct{} `example:"/test"`

	Body struct {
		Name string
	} `content-type:"application/json, application/xml"`

	Response struct {
		ID int
	} `content-type:"application/json, application/xml"`
}

func (r *MultiMimeRequest) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error { return nil }
func (r *MultiMimeRequest) EncodeResponse(ctx context.Context, out http.ResponseWriter, obj interface{}) {}
func (r *MultiMimeRequest) Handle(ctx context.Context, w http.ResponseWriter) error { return nil }

func TestMultipleMimeTypes(t *testing.T) {
	router := chi.NewRouter()
	infos := openapi3.Info{
		Title:   "test api",
		Version: "1.0.0",
	}

	b, err := New(router, &infos)
	require.NoError(t, err)

	err = b.Get(router, "/test", &MultiMimeRequest{})
	require.NoError(t, err)

	swagger, err := b.GenerateSwagger(context.Background(), shared.NewChipiCallbacks(nil))
	require.NoError(t, err)

	op := swagger.Paths.Find("/test").Get
	require.NotNil(t, op)

	require.NotNil(t, op.RequestBody.Value)
	require.NotNil(t, op.RequestBody.Value.Content["application/json"])
	require.NotNil(t, op.RequestBody.Value.Content["application/xml"])

	require.NotNil(t, op.Responses.Map()["200"].Value)
	require.NotNil(t, op.Responses.Map()["200"].Value.Content["application/json"])
	require.NotNil(t, op.Responses.Map()["200"].Value.Content["application/xml"])
}
