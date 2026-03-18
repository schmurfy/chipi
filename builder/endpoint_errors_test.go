package builder

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/ruwanego/chipi/shared"
	"github.com/stretchr/testify/require"
)

type EndpointSpecificError struct {
	Details string
}

type MyEndpointErrorRequest struct {
	Path struct{} `example:"/test"`

	Errors struct {
		BadRequest          EndpointSpecificError `chipi:"400" description:"Specific Bad Request"`
		InternalServerError struct{}              `chipi:"500" description:"Internal Server Error"`
	}
}

func (r *MyEndpointErrorRequest) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error { return nil }
func (r *MyEndpointErrorRequest) EncodeResponse(ctx context.Context, out http.ResponseWriter, obj interface{}) {}
func (r *MyEndpointErrorRequest) Handle(ctx context.Context, w http.ResponseWriter) error { return nil }

func TestEndpointErrors(t *testing.T) {
	router := chi.NewRouter()
	infos := openapi3.Info{
		Title:   "test api",
		Version: "1.0.0",
	}

	b, err := New(router, &infos)
	require.NoError(t, err)

	b.AddGlobalResponse(400, "Global Bad Request", MyGlobalError{}, "application/json")

	err = b.Get(router, "/test", &MyEndpointErrorRequest{})
	require.NoError(t, err)

	swagger, err := b.GenerateSwagger(context.Background(), shared.NewChipiCallbacks(nil))
	require.NoError(t, err)

	op := swagger.Paths.Find("/test").Get
	require.NotNil(t, op)

	// Endpoint specific error should override the global error
	require.NotNil(t, op.Responses.Value("400"))
	require.Equal(t, "Specific Bad Request", *op.Responses.Value("400").Value.Description)
	require.NotNil(t, op.Responses.Value("400").Value.Content["application/json"])

	require.NotNil(t, op.Responses.Value("500"))
	require.Equal(t, "Internal Server Error", *op.Responses.Value("500").Value.Description)
	require.NotNil(t, op.Responses.Value("500").Value.Content["application/json"])
}
