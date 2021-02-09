package main

import (
	"io"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	// _ "embed"

	"github.com/schmurfy/chipi"
)

// go:embed index.html
var indexFile []byte

func main() {
	api, err := chipi.New(&openapi3.Info{
		Title:       "test api",
		Description: "a great api",
	})

	api.AddServer(&openapi3.Server{
		URL: "http://127.0.0.1:2121",
	})

	api.AddSecurityRequirement(openapi3.SecurityRequirement{
		"api_key": []string{},
	})

	api.AddSecurityScheme("api_key", &openapi3.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		In:           "header",
	})

	if err != nil {
		panic(err)
	}
	router := chi.NewRouter()

	router.Use(cors.AllowAll().Handler)

	router.Get("/doc.json", api.ServeSchema)

	router.Get("/doc", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(w, f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// embed
		// w.Write(indexFile)
	})

	router.Group(func(r chi.Router) {
		// TODO: add options for deprecated, tags
		err := api.Get(r, "/pet/{Id}", &GetPetRequest{})
		if err != nil {
			panic(err)
		}

		err = api.Post(r, "/pet", &CreatePetRequest{})
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

		err = api.Get(r, "/user/{Name}", &DownloadResumeRequest{})
		if err != nil {
			panic(err)
		}

	})

	http.ListenAndServe(":2121", router)
}
