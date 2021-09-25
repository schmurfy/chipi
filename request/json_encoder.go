package request

import (
	"encoding/json"
	"net/http"
)

type JsonResponseEncoder struct{}

func (e *JsonResponseEncoder) EncodeResponse(w http.ResponseWriter, obj interface{}) {
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
