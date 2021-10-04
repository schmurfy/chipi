package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type User struct {
	Name string `json:"name"`
	Pets []Pet  `json:"pets"`
}

// @summary
// get a user
type GetUserRequest struct {
	Path struct {
		// @description
		// # some title
		// The _name_ of the __user__
		// ***
		// - one
		// - two
		// ```
		// quoted text
		// ````
		Name string `example:"roger"`
		Id   int32
	} `example:"/user/clark"`

	Query struct{}

	// @description
	// the returned user
	Response User
}

func (r *GetUserRequest) Handle(ctx context.Context, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&User{
		Name: r.Path.Name,
		Pets: []Pet{
			{Id: 3, Name: "Rex"},
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// @summary
// upload user's resume)
type UploadResumeRequest struct {
	Path struct {
		Name string `example:"john"`
	} `example:"/user/john"`
	Query struct{}

	Body struct {
		File1 []byte
		File2 []byte
	} `content-type:"multipart/form-data"`

	// @description
	// returns nothing
	Response struct{}
}

func (r *UploadResumeRequest) Handle(ctx context.Context, w http.ResponseWriter) {

}

// @summary
// download user's resume
type DownloadResumeRequest struct {
	Path struct {
		Name string `example:"john"`
	} `example:"/user/john/download"`

	Query struct{}

	// @description
	// the resume
	Response []byte `description:"the user resume, maybe ?"`
}

func (r *DownloadResumeRequest) Handle(ctx context.Context, w http.ResponseWriter) {
	data := []byte("some data")
	reader := bytes.NewReader(data)

	w.Header().Set("Content-Type", "image/jpeg")

	_, err := io.Copy(w, reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
