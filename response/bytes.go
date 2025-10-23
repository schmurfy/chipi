package response

import (
	"context"
	"net/http"
)

type BytesEncoder struct{}

func (e *BytesEncoder) EncodeResponse(ctx context.Context, w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/octet-stream")

	raw, ok := obj.([]byte)
	if !ok {
		http.Error(w, "response type invalid", http.StatusBadRequest)
		return
	}

	_, err := w.Write(raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
