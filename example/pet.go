package main

import (
	"encoding/json"
	"net/http"
)

type Pet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetPetRequest struct {
	Path struct {
		Id string `example:"42"`
	} `example:"/pet/5"`

	Response Pet
}

func GetPet(w http.ResponseWriter, req interface{}) {
	r := req.(*GetPetRequest)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&Pet{
		Id:   r.Path.Id,
		Name: "Fido",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type CreatePetRequest struct {
	Path     struct{} `example:"/pet"`
	Body     Pet
	Response Pet
}

func CreatePet(w http.ResponseWriter, req interface{}) {
	r := req.(*CreatePetRequest)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
