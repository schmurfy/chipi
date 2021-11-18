package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/schmurfy/chipi/response"
)

type RequestWithFields struct {
	Fields []string
}

func (d *RequestWithFields) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&target)
	if err != nil {
		return err
	}

	d.Fields = []string{"one", "two"}
	return nil
}

type CreateBikeRequest struct {
	response.ErrorEncoder
	RequestWithFields

	Path struct{} `example:"/bikes/"`

	Body struct {
		Id string
	}
}

func (r *CreateBikeRequest) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	err := r.RequestWithFields.DecodeBody(body, target, obj)
	if err != nil {
		return err
	}

	fmt.Printf("%x %x\n", r, obj)

	r.Fields = []string{r.Body.Id}

	return nil
}

func (r *CreateBikeRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	fmt.Printf("req: %+v\n", r)
	return json.NewEncoder(w).Encode(r)
}
