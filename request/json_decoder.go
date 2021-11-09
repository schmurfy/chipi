package request

import (
	"encoding/json"
	"io"
)

type JsonBodyDecoder struct{}

func (d *JsonBodyDecoder) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	// otherwise use the default decoder
	decoder := json.NewDecoder(body)
	return decoder.Decode(&target)
}
