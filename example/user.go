package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/schmurfy/chipi/response"
)

type User struct {
	Name string `json:"name"`
	Pets []Pet  `json:"pets"`
}

// @summary
// get a user
type GetUserRequest struct {
	response.ErrorEncoder
	response.JsonEncoder

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

func (r *GetUserRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(&User{
		Name: r.Path.Name,
		Pets: []Pet{
			{Id: 3, Name: "Rex"},
		},
	})
}

// @summary
// upload user's resume)
type UploadResumeRequest struct {
	response.ErrorEncoder
	response.JsonEncoder

	Path struct {
		Name string `example:"john"`
	} `example:"/user/john"`
	Query struct{}

	Body struct {
		File1 []byte `json:"file1"`
		File2 []byte
	} `content-type:"multipart/form-data"`

	// @description
	// returns nothing
	Response struct{}
}

func (r *UploadResumeRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	return nil
}

// @summary
// download user's resume
type DownloadResumeRequest struct {
	response.ErrorEncoder
	response.JsonEncoder

	Path struct {
		Name string `example:"john"`
	} `example:"/user/john/download"`

	Query struct{}

	// @description
	// the resume
	Response []byte `description:"the user resume, maybe ?"`
}

func (r *DownloadResumeRequest) Handle(ctx context.Context, w http.ResponseWriter) error {
	data := []byte("some data")
	reader := bytes.NewReader(data)

	w.Header().Set("Content-Type", "image/jpeg")

	_, err := io.Copy(w, reader)
	return err
}
