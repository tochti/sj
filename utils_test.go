package sj

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_ParseSeriesRequest_OK(t *testing.T) {
	data := `
	{
		"Data": {
			"Title": "Title",
			"Image": "Image"
		}
	}`

	body := bytes.NewReader([]byte(data))
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}

	ginCtx := gin.Context{}
	ginCtx.Request = req

	s, err := ParseNewSeriesRequest(&ginCtx)

	expect := Series{
		Title: "Title",
		Image: "Image",
	}

	if err := EqualSeries(expect, s); err != nil {
		t.Fatal(err)
	}

}

func Test_ParseUserRequest_OK(t *testing.T) {
	data := `
	{
		"Data": {
			"Name": "name",
			"Password": "password"
		}
	}`

	body := bytes.NewReader([]byte(data))
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}

	ginCtx := gin.Context{}
	ginCtx.Request = req

	u, err := ParseNewUserRequest(&ginCtx)

	expect := User{
		Name:     "name",
		Password: "password",
	}

	if err := EqualUser(expect, u); err != nil {
		t.Fatal(err)
	}

}

func Test_ParseAppendSeriesListRequest_OK(t *testing.T) {
	data := `
	{
		"Data": {
			"UserID": 1,
			"SeriesID": 2
		}
	}`

	body := bytes.NewReader([]byte(data))
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}

	ginCtx := gin.Context{}
	ginCtx.Request = req

	s, err := ParseAppendSeriesListRequest(&ginCtx)
	if err != nil {
		t.Fatal(err)
	}

	expect := SeriesListRequestData{
		UserID:   1,
		SeriesID: 2,
	}

	if expect.UserID != s.UserID ||
		expect.SeriesID != s.SeriesID {
		m := fmt.Sprintf("Expect %v was %v", expect, s)
		t.Fatal(m)
	}

}
