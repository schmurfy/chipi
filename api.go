package chipi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/builder"
)

func New(infos *openapi3.Info) (*builder.Builder, error) {
	return builder.New(infos)
}
