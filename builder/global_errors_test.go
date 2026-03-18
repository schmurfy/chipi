package builder

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/shared"
	"github.com/stretchr/testify/require"
)

type MyGlobalError struct {
	Message string
}

func TestGlobalErrors(t *testing.T) {
	router := chi.NewRouter()
	infos := openapi3.Info{
		Title:   "test api",
		Version: "1.0.0",
	}

	b, err := New(router, &infos)
	require.NoError(t, err)

	b.AddGlobalResponse(400, "Bad Request", MyGlobalError{}, "application/json")
	b.AddGlobalResponse(500, "Internal Server Error", nil, "")

	err = b.Get(router, "/test", &MultiMimeRequest{})
	require.NoError(t, err)

	swagger, err := b.GenerateSwagger(context.Background(), shared.NewChipiCallbacks(nil))
	require.NoError(t, err)

	op := swagger.Paths.Find("/test").Get
	require.NotNil(t, op)

	require.NotNil(t, op.Responses.Value("400"))
	require.Equal(t, "Bad Request", *op.Responses.Value("400").Value.Description)
	require.NotNil(t, op.Responses.Value("400").Value.Content["application/json"])

	require.NotNil(t, op.Responses.Value("500"))
	require.Equal(t, "Internal Server Error", *op.Responses.Value("500").Value.Description)
	require.Nil(t, op.Responses.Value("500").Value.Content["application/json"])
}
