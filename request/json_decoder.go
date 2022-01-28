package request

import (
	"encoding/json"
	"io"
)

type JsonBodyDecoder struct{}

func (d *JsonBodyDecoder) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	// otherwise use the default decoder
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&target)
	if err != nil && err.Error() == "EOF" {
		return nil
	}
	return err
}
