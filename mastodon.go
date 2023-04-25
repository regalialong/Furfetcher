package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// https://stackoverflow.com/a/56696333 - Amended without Testing
func createMultipartFormData(fieldName, fileName string) (bytes.Buffer, *multipart.Writer) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer
	file := mustOpen(fileName)
	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		panic(err)
	}
	w.Close()
	return b, w
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		pwd, _ := os.Getwd()
		fmt.Println("PWD: ", pwd)
		panic(err)
	}
	return r
}

func send_file(filename string) (response *MediaResponse) {
	CONSTRUCTED := configuration.BASEURL + "/api/v2/media"
	b, w := createMultipartFormData("file", filename)

	r, err, client := buildRequest(CONSTRUCTED, &b)
	r.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		panic(res.Status)
	}
	contents := &MediaResponse{}
	derr := json.NewDecoder(res.Body).Decode(contents)
	if derr != nil {
		panic(derr)
	}
	return contents
}

func status(parameters StatusParameters) (response Status) {
	CONSTRUCTED := configuration.BASEURL + "/api/v1/statuses"

	body, err := json.Marshal(parameters)
	if err != nil {
		panic(err)
	}

	r, err, client := buildRequest(CONSTRUCTED, bytes.NewBuffer(body))
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	post := &Status{}
	derr := json.NewDecoder(res.Body).Decode(post)
	if derr != nil {
		panic(derr)
	}

	if res.StatusCode != http.StatusOK {
		panic(res.Status)
	}

	return response
}

func buildRequest(url string, body *bytes.Buffer) (*http.Request, error, *http.Client) {
	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/json")
	r.SetBasicAuth(configuration.USERNAME, configuration.PASSWORD)

	client := &http.Client{}
	return r, err, client
}
