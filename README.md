
After being frustrated multiple times about the lack of easy way to generate an openapi doc directly from
the code I created this library as an experiment and it went way further than I expected.

## Other solutions

My main problem with the alternatives is simple: I don't want to maintain comments to describe my apis, in my experience those will slowly drift and become inaccurate. On the other hand if the code itself is the documentation it cannot technically drift or else it will not longer works.

## My solution

The library is based on `chi` which is, so far, the best http router I found.

Each api endpoint is described by a structure:

```go
type GetPetRequest struct {
	Path struct {
		// @description
		// You can add multiline _markdown_ description
		//
		Id int32 `example:"42"`
	} `example:"/pet/5"`

	Query struct {
		Count *int `example:"2"`
	}

	Response Pet
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


## Supported OpenAPI (v3) attributes

### Path

[reference](https://swagger.io/specification/#parameter-object)

- description [comment,tag]
- example [comment,tag]
- deprecated [tag]
- style [tag]
- explode [tag]

### Query

[reference](https://swagger.io/specification/#parameter-object)

( same as path parameters )

### Header

[reference](https://swagger.io/specification/#header-object)

( same as path parameters )

### Body

[reference](https://swagger.io/specification/#body-object)

- description [comment,tag]
- required [tag]
- content-type [tag]

### Response

[reference](https://swagger.io/specification/#response-object)

- description [comment,tag]

## Caveats

This solution is not perfect and lack some features but I am sure a way to implement them can be found if needed:

- no way to specify errors response, in my experience errors are often reported in a similar way for the whole api which may be documented as an introduction to the api.

- no way to specify multiple mime type for body/response: that is a choice but what I need is a simple solution, I am not trying to solve every problems.

