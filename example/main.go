package main

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	"github.com/schmurfy/chipi"
)

func main() {
	api, err := chipi.New(&openapi3.Info{
		Title:       "test api",
		Description: "a great api",
	})

	api.AddServer(&openapi3.Server{
		URL: "http://127.0.0.1:2121",
	})

	if err != nil {
		panic(err)
	}
	router := chi.NewRouter()

	router.Use(cors.AllowAll().Handler)

	router.Get("/doc.json", api.ServeSchema)

	router.Group(func(r chi.Router) {
		api.Get(r, "/pet/{Id}", GetPetRequest{}, GetPet)
		api.Post(r, "/pet", CreatePetRequest{}, CreatePet)
	})

	http.ListenAndServe(":2121", router)
}
