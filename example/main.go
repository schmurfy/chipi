package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	_ "embed"

	"github.com/schmurfy/chipi"
)

//go:embed index.html
var indexFile []byte

//go:embed redoc.html
var redocFile []byte

func main() {
	router := chi.NewRouter()

	api, err := chipi.New(router, &openapi3.Info{
		Title:       "test api",
		Description: "a great api",
		License: &openapi3.License{
			Name: "MIT",
			URL:  "https://raw.githubusercontent.com/schmurfy/chipi/master/LICENSE",
		},
	})

	api.AddServer(&openapi3.Server{
		URL: "http://127.0.0.1:2121",
	})

	api.AddServer(&openapi3.Server{
		URL: "http://127.0.0.1:2122",
	})

	// https://spec.openapis.org/oas/latest.html#security-scheme-object
	// api.AddSecurityRequirement(openapi3.SecurityRequirement{
	// 	"api_key": []string{},
	// 	"basic":   []string{},
	// })

	// api.AddSecurityScheme("api_key", &openapi3.SecurityScheme{
	// 	Type:         "http",
	// 	Scheme:       "bearer",
	// 	BearerFormat: "JWT",
	// 	In:           "header",
	// })

	api.AddSecurityScheme("basic", &openapi3.SecurityScheme{
		Type:   "http",
		Scheme: "basic",
		In:     "header",
	})

	if err != nil {
		panic(err)
	}
	router.Use(cors.AllowAll().Handler)

	router.Get("/doc.json", api.ServeSchema)
	router.Get("/doc2.json", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("/tmp/doc.json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer f.Close()

		_, err = io.Copy(w, f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	router.Get("/redoc", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(redocFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.Get("/doc", func(w http.ResponseWriter, r *http.Request) {
		// embed
		_, err := w.Write(indexFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	r := router.Group(func(r chi.Router) {

		petRouter := chi.NewRouter()
		r.Mount("/pet", petRouter)

		err := api.Get(petRouter, "/{Id}", &GetPetRequest{})
		if err != nil {
			log.Fatalf("%+v", err)
		}

		err = api.Post(petRouter, "/", &CreatePetRequest{})
		if err != nil {
			panic(err)
		}

		err = api.Get(r, "/user/{Name}", &GetUserRequest{})
		if err != nil {
			panic(err)
		}

		err = api.Post(r, "/user/{Name}", &UploadResumeRequest{})
		if err != nil {
			panic(err)
		}

	})

	err = api.Get(r, "/user/{Name}/download", &DownloadResumeRequest{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Started on 127.0.0.1:2121\n")

	err = http.ListenAndServe(":2121", router)
	if err != nil {
		panic(err)
	}
}
