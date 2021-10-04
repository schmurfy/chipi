[![codecov](https://codecov.io/gh/schmurfy/chipi/branch/master/graph/badge.svg?token=A6413R1ZXH)](https://codecov.io/gh/schmurfy/chipi)
[![Go Report Card](https://goreportcard.com/badge/github.com/schmurfy/chipi)](https://goreportcard.com/report/github.com/schmurfy/chipi)

After being frustrated multiple times about the lack of easy way to generate an openapi doc directly from
the code I created this library as an experiment and it went way further than I expected.

## Other solutions

My main problem with the alternatives is simple: I don't want to maintain comments to describe my apis, in my experience those will slowly drift and become inaccurate. On the other hand if the code itself is the documentation it cannot technically drift or else it will not longer works.

## My solution

The library is based on `chi` which is, so far, the best http router I found.

Each api endpoint is described by a structure:

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
```

And you can use it like that:

```go
err := api.Get(r, "/pet/{Id}", &GetPetRequest{})
if err != nil {
  panic(err)
}
```

- `Path` is mandatory and describe the path parameters
- `Query` is optional and will match query parameters (ex: "?count=4")
- `Body` is optional and if present can be either a structure (json tags will be honored)
- `Response` is also optional and define what is returned when eveything works well


## Supported OpenAPI (v3.1) attributes

### Structures

Special tags can be used on structure's fields to set specific behaviors:

- ignored: the field will not show at all, triggered by:
  - `json:"-"`
  - `chipi:"ignore"`
- read only: field only valid on read
  - `chipi:"readonly"`
- write only: field only valid on write
  - `chipi:"writeonly"`
- nullable: the field can be set to `null`
  - `chipi:"nullable"`
- deprecated
  - `chipi:"deprecated"`

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

## Caveats

This solution is not perfect and lack some features but I am sure a way to implement them can be found if needed:

- no way to specify errors response, in my experience errors are often reported in a similar way for the whole api which may be documented as an introduction to the api.

- no way to specify multiple mime type for body/response: that is a choice but what I need is a simple solution, I am not trying to solve every problems.

