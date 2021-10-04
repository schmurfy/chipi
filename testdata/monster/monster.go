package monster

type Monster struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

// @tag
// monster
//
// @summary
// Grab a monster and bring it to you
// knowing its Id
//
// @deprecated
type GetMonsterRequest struct {
	Path struct {
		// @description
		// The _Id_ of the monster you want to
		// fetch
		Id int32
	} `example:"/monster/3"`

	// @description
	// the query
	Query struct {
		// @description
		// If true the request will block until
		// the monster was actually created
		// @example
		// ahhhhhh !
		Blocking bool
	}

	Header struct {
		// @description
		// The _ApiKey_ is required to
		// check for authorization
		ApiKey string

		// @description
		// This may be important
		Something string
	}

	// @description
	// what is returned
	Response Monster
}
