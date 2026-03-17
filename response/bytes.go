package response

import (
	"context"
	"fmt"
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
		fmt.Printf("failed to write response: %v\n", err)
		return
	}
}
