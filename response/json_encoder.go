package response

import (
	"encoding/json"
	"net/http"
)

type JsonEncoder struct{}

func (e *JsonEncoder) EncodeResponse(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
