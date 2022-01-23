package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/schmurfy/chipi/response"
)

type Pet struct {
	Id           int32  `json:"id"`
	Name         string `json:"name"`
	Count        *int   `json:"count"`
	User         *User  `json:"user" chipi:"nullable,deprecated"`
	ReadOnly     int    `json:"red_only" chipi:"readonly"`
	IgnoreString string `chipi:"ignore"`
}

type Location struct {
	Type        string
	Coordinates []float64
}

// @tag
// pets
//
// @summary
// fetch a pet
//
// @deprecated
type GetPetRequest struct {
	response.ErrorEncoder
	response.JsonEncoder

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
		Count    *int     `example:"2" description:"it counts... something ?"`
		Age      []int    `example:"[1,3,4]" style:"form" explode:"false" description:"line one\nline two" chipi:"required"`
		Names    []string `example:"[\"a\",\"b\",\"c\"]" style:"form" explode:"false" description:"line one\nline two"`
		OldField string   `chipi:"deprecated"`

		// @example
		// {"type": "point", "coordinates": [0.2, 9.0]}
		//
		// @description
		// # first
		// the location near the pet
		// ## second
		// some list of things:
		// - one
		// - two
		Location *Location
	}

	Header struct {
		ApiKey string
	}

	// @description
	// the returned pet
	Response Pet
}

func (r *GetPetRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&Pet{
		Id:    r.Path.Id,
		Name:  "Fido",
		Count: r.Query.Count,
	})

	fmt.Printf("names: %+v\n", r.Query.Names)
	fmt.Printf("location: %+v\n", r.Query.Location)

	return err
}

// @tag
// pets
// @summary
// add a new pet
type CreatePetRequest struct {
	response.ErrorEncoder
	response.JsonEncoder

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
	Body *Pet

	// @description
	// the returned pet
	Response Pet
}

func (r *CreatePetRequest) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	pet := target.(*Pet)

	c := 56
	pet.Count = &c
	return nil
}

func (r *CreatePetRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(r.Body)
}
