package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type JsonEncoder struct{}

func (e *JsonEncoder) EncodeResponse(ctx context.Context, w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		fmt.Printf("failed to write response: %v\n", err)
		return
	}

}
