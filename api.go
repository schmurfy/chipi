package chipi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/builder"
)

func New(r *chi.Mux, infos *openapi3.Info) (*builder.Builder, error) {
	return builder.New(r, infos)
}
