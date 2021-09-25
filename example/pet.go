package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Pet struct {
	Id    int32  `json:"id"`
	Name  string `json:"name"`
	Count *int   `json:"count"`
	User  *User  `json:"user"`
}

type GetPetRequest struct {
	Path struct {
		// @description
		// Id is so great !
		// yeah !!
		//
		// @example
		// 789
		Id int32
	} `example:"/pet/5"`

	Query struct {
		Count *int     `example:"2" deprecated:"true" description:"it counts... something ?"`
		Age   []int    `example:"[1,3,4]" style:"form" explode:"false" description:"line one\nline two"`
		Names []string `example:"[\"a\",\"b\",\"c\"]" style:"form" explode:"false" description:"line one\nline two"`
	}

	Header struct {
		ApiKey string
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

	fmt.Printf("names: %+v\n", r.Query.Names)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type CreatePetRequest struct {
	// @description
	// this is a wonderful path with
	// a lot of things inside, really a great path !
	Path  struct{} `example:"/pet"`
	Query struct{}

	/* @description
	some comment
	how cool is that ?
	another line
	and more !
	*/
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
