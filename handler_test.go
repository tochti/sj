package sj

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/gin-gonic/gin"
)

type (
	TestRequest struct {
		Body    string
		Handler http.Handler
		Header  http.Header
	}
)

func (t *TestRequest) SendWithToken(method, path, token string) *httptest.ResponseRecorder {
	reqData := *t
	body := bytes.NewBufferString(reqData.Body)
	reqData.Header.Add("X-XSRF-TOKEN", token)

	req, _ := http.NewRequest(method, path, body)
	req.Header = reqData.Header
	w := httptest.NewRecorder()
	reqData.Handler.ServeHTTP(w, req)
	*t = reqData
	return w
}

func (t *TestRequest) Send(method, path string) *httptest.ResponseRecorder {
	reqData := *t
	body := bytes.NewBufferString(reqData.Body)

	req, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	reqData.Handler.ServeHTTP(w, req)
	*t = reqData
	return w
}

func EqualResponse(expect interface{}, result *bytes.Buffer) error {
	json, err := json.Marshal(expect)
	if err != nil {
		return err
	}

	t1 := bytes.Trim(json, "\n")
	t2 := bytes.Trim(result.Bytes(), "\n")

	if !bytes.Equal(t1, t2) {
		m := fmt.Sprintf("Expect %v was %v", string(t1), string(t2))
		return errors.New(m)
	}

	return nil
}

func Test_POST_User_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	m := "SELECT ID,Name,Password FROM %v"
	q := fmt.Sprintf(m, UserTable)
	mock.ExpectQuery(q).WillReturnError(sql.ErrNoRows)

	q = fmt.Sprintf("INSERT INTO %v", UserTable)
	mock.ExpectExec(q).
		WithArgs("devilXX", NewSha512Password("123")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	app := AppCtx{
		DB: db,
	}
	srv := gin.New()
	srv.POST("/", NewAppHandler(app, NewUserHandler))

	body := `
	{
		"Data": {
			"Name": "devilXX",
			"Password": "123"
		}
	}
	`

	req := TestRequest{
		Body:    body,
		Handler: srv,
		Header:  http.Header{},
	}
	resp := req.Send("POST", "/")

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	u := User{
		Name: "devilXX",
	}
	expect := NewSuccessResponse(u)
	err = EqualResponse(expect, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_PATCH_SeriesList_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)
	seriesID := int64(2)

	q := fmt.Sprintf("INSERT INTO %v", SeriesListTable)
	mock.ExpectExec(q).
		WithArgs(userID, seriesID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	app := AppCtx{
		DB: db,
	}
	srv := gin.New()
	srv.POST("/", NewAppHandler(app, AppendSeriesListHandler))

	body := `
	{
		"Data": {
			"UserID": 1, 
			"SeriesID": 2
		}
	}
	`

	req := TestRequest{
		Body:    body,
		Handler: srv,
		Header:  http.Header{},
	}
	resp := req.Send("POST", "/")

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	expect := NewSuccessResponse("")
	err = EqualResponse(expect, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}
