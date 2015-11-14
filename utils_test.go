package sj

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

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
		SeriesID: 2,
	}

	if expect.SeriesID != s.SeriesID {
		m := fmt.Sprintf("Expect %v was %v", expect, s)
		t.Fatal(m)
	}

}

func Test_ParseUpdateLastWatchedRequest_OK(t *testing.T) {
	data := `
	{
		"Data": {
			"UserID": 1,
			"SeriesID": 2,
			"Session": 3,
			"Episode": 4
		}
	}`

	body := bytes.NewReader([]byte(data))
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}

	ginCtx := gin.Context{}
	ginCtx.Request = req

	s, err := ParseUpdateLastWatchedRequest(&ginCtx)
	if err != nil {
		t.Fatal(err)
	}

	expect := LastWatched{
		SeriesID: 2,
		Session:  3,
		Episode:  4,
	}

	if expect.SeriesID != s.SeriesID ||
		expect.Session != s.Session ||
		expect.Episode != s.Episode {
		m := fmt.Sprintf("Expect %v was %v", expect, s)
		t.Fatal(m)
	}

}

func Test_SaveImage_OK(t *testing.T) {

	img := []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73,
		72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0, 144,
		119, 83, 222, 0, 0, 0, 12, 73, 68, 65, 84, 8, 215, 99, 184,
		120, 241, 34, 0, 4, 234, 2, 116, 26, 41, 186, 204, 0, 0, 0,
		0, 73, 69, 78, 68, 174, 66, 96, 130}

	name, err := ioutil.TempDir(".", "test")
	tmpDir := path.Join(".", name)
	defer os.RemoveAll(tmpDir)

	err = ioutil.WriteFile(path.Join(tmpDir, "test.png"), img, 0755)
	if err != nil {
		t.Fatal(err)
	}

	srv := gin.New()
	srv.Static("/img", tmpDir)
	srvAddr := "127.0.0.1:63000"
	go func() {
		srv.Run(srvAddr)
	}()

	time.Sleep(500 * time.Millisecond)
	url := fmt.Sprintf("http://%v/img/%v", srvAddr, "test.png")
	name, err = SaveImage(url, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	c, err := ioutil.ReadFile(path.Join(tmpDir, name))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(img, c) {
		t.Fatal("Expect", img, "was", c)
	}

}
