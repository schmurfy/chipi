[![codecov](https://codecov.io/gh/schmurfy/chipi/branch/master/graph/badge.svg?token=A6413R1ZXH)](https://codecov.io/gh/schmurfy/chipi)
[![Go Report Card](https://goreportcard.com/badge/github.com/schmurfy/chipi)](https://goreportcard.com/report/github.com/schmurfy/chipi)

Chipi is a simple, code-driven OpenAPI v3.1 generator for the [`chi`](https://github.com/go-chi/chi) HTTP router.

After being frustrated multiple times about the lack of easy way to generate an OpenAPI doc directly from
the code, I created this library as an experiment and it went way further than I expected.

## Why Chipi?

My main problem with the alternatives is simple: I don't want to maintain separate comments to describe my apis. In my experience those will slowly drift and become inaccurate. On the other hand, if the code itself is the documentation, it cannot technically drift or else it will no longer work.

With Chipi, you write strongly-typed request/response structs, register them using a wrapper, and your API documentation is generated dynamically.

## Installation

```bash
go get github.com/schmurfy/chipi
```

## Getting Started

To use Chipi, you wrap your `chi` router with the `chipi.Builder`. You define your endpoints using a struct that satisfies the `HandlerInterface` interface, and Chipi automatically reflects on your struct to build an OpenAPI 3.1 definition.

### 1. Initialize the Router and Builder

```go
package main

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/builder"
	"github.com/schmurfy/chipi/shared"
)

func main() {
	r := chi.NewRouter()

	infos := &openapi3.Info{
		Title:   "My Awesome API",
		Version: "1.0.0",
	}

	// Initialize the Chipi builder
	api, err := builder.New(r, infos)
	if err != nil {
		panic(err)
	}

	// Optional: Add global API error models
	api.AddGlobalResponse(500, "Internal Server Error", MyErrorModel{}, "application/json")

	// Register your endpoints using the builder
	err = api.Get(r, "/pet/{Id}", &GetPetRequest{})
	if err != nil {
		panic(err)
	}

	// Expose the OpenAPI schema as JSON
	r.Get("/openapi.json", api.ServeSchema)

	http.ListenAndServe(":8080", r)
}

type MyErrorModel struct {
	Message string `json:"message"`
}
```

### 2. Define an Endpoint using a Request Struct

Each API endpoint is described by a structure. This structure must implement the `Handle(context.Context, http.ResponseWriter) error` interface.

```go
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

	Body struct {
		Name string
	} `content-type:"application/json, application/xml"` // Support multiple MIME types via comma separation

	// @description
	// the returned pet
	Response Pet `content-type:"application/json"`

	Errors struct {
		BadRequest MyErrorModel `chipi:"400" description:"Specific Bad Request"`
	}
}

func (r *GetPetRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&Pet{
		Id:    r.Path.Id,
		Name:  "Fido",
		Count: r.Query.Count,
	})

	return err
}

type Pet struct {
	Id    int32
	Name  string
	Count *int
}

type Location struct {
	Type        string
	Coordinates []float64
}
```

### Anatomy of a Request Struct

- `Path`: **Mandatory**. Describes the path parameters. Must have an `example` tag matching your chi route parameter.
- `Query`: Optional. Matches query parameters (e.g. `?count=4`).
- `Header`: Optional. Matches headers sent in the request.
- `Body`: Optional. Used for parsing incoming request bodies. Requires your struct to implement `wrapper.BodyDecoder` if used. Can be tagged with `content-type` to define single or multiple comma-separated MIME types.
- `Response`: Optional. Defines what is returned when the endpoint is successful. Requires your struct to implement `wrapper.ResponseEncoder` if used. Can be tagged with `content-type`.
- `Errors`: Optional. Defines endpoint-specific errors. Uses the `chipi` struct tag to indicate the HTTP Status Code (e.g. `chipi:"400"`). These override any Global Errors configured on the builder for the same status code.


## Supported OpenAPI (v3.1) attributes

### Structures

Special tags can be used on a structure's fields to set specific behaviors:

- **ignored**: The field will not show at all, triggered by:
  - `json:"-"`
  - `chipi:"ignore"`
- **read only**: Field only valid on read
  - `chipi:"readonly"`
- **write only**: Field only valid on write
  - `chipi:"writeonly"`
- **nullable**: The field can be set to `null`
  - `chipi:"nullable"`
- **required**: Marks the field as strictly required
  - `chipi:"required"`
- **deprecated**: Marks the field as deprecated
  - `chipi:"deprecated"`
- **example**: Generates an example in the OpenAPI spec
  - `example:"field example"`
- **description**: Generates a description for the field
  - `description:"field description"`
- **content-type**: Defines the MIME type(s). You can specify a comma-separated list of MIME types on `Body` and `Response` fields (e.g. `content-type:"application/json, application/xml"`).

### Path

[reference](https://spec.openapis.org/oas/v3.1.0.html#parameter-object)

- example [comment,tag]
- description [comment,tag]
- style [tag]
- explode [tag]
- deprecated [chipi-tag]

### Query

[reference](https://spec.openapis.org/oas/v3.1.0.html#parameter-object)

( same as path parameters )
- required [chipi-tag]

### Header

[reference](https://spec.openapis.org/oas/v3.1.0.html#parameter-object)

( same as path parameters )
- required [chipi-tag]

### Body

[reference](https://spec.openapis.org/oas/v3.1.0.html#request-body-object)

- content-type [tag]
- description [comment,tag]
- required [chipi-tag]

### Response

[reference](https://spec.openapis.org/oas/v3.1.0.html#response-object)

- description [comment,tag]
- content-type [tag]
