package response

import (
	"context"
	"net/http"
)

type ErrorEncoder struct{}

func (e *ErrorEncoder) HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
