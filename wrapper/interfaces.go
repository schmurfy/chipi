package wrapper

import (
	"context"
	"io"
	"net/http"
)

// BodyDecoder is required for structures with a `Body` field
type BodyDecoder interface {
	DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error
}

// ResponseEncoder is required for structures with a `Response` field
type ResponseEncoder interface {
	EncodeResponse(out http.ResponseWriter, obj interface{})
}

type HandlerInterface interface {
	Handle(context.Context, http.ResponseWriter) error
}

type ErrorHandlerInterface interface {
	HandleError(context.Context, http.ResponseWriter, error)
}

type HandlerWithRequestInterface interface {
	Handle(context.Context, *http.Request, http.ResponseWriter) error
}
