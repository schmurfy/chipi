package main

import (
	"encoding/json"
	"net/http"
)

type Pet struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type GetPetRequest struct {
	Path struct {
		Id string `example:"42"`
	} `example:"/pet/5"`

	Query struct {
		Count int `example:"2"`
	}

	Response Pet
}

func (r *GetPetRequest) Handle(w http.ResponseWriter) {
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
	Body     Pet
	Response Pet
}

func (r *CreatePetRequest) Handle(w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
