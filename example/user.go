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

type GetUserRequest struct {
	Path struct {
		Name string `example:"john"`
	} `example:"/user/john"`

	Query struct{}

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

type UploadResumeRequest struct {
	Path struct {
		Name string `example:"john"`
	} `example:"/user/john"`
	Query struct{}

	Body struct {
		File1 []byte
		File2 []byte
	} `content-type:"multipart/form-data"`
}

func (r *UploadResumeRequest) Handle(ctx context.Context, w http.ResponseWriter) {

}

type DownloadResumeRequest struct {
	Path struct {
		Name string `example:"john"`
	} `example:"/user/john"`

	Query struct{}

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
