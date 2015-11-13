package sj

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_ParseSeries_OK(t *testing.T) {
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
