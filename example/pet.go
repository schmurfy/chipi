package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Pet struct {
	Id    int32  `json:"id"`
	Name  string `json:"name"`
	Count *int   `json:"count"`
}

type GetPetRequest struct {
	Path struct {
		Id int32 `example:"42"`
	} `example:"/pet/5"`

	Query struct {
		Count *int `example:"2"`
	}

	Response Pet
}

func (r *GetPetRequest) Handle(ctx context.Context, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&Pet{
		Id:    r.Path.Id,
		Name:  "Fido",
		Count: r.Query.Count,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type CreatePetRequest struct {
	Path     struct{} `example:"/pet"`
	Query    struct{}
	Body     *Pet
	Response Pet
}

func (r *CreatePetRequest) DecodeBody(body io.ReadCloser, target interface{}) error {
	pet := target.(*Pet)

	c := 56
	pet.Count = &c
	return nil
}

func (r *CreatePetRequest) Handle(ctx context.Context, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
