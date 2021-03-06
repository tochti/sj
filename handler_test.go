package sj

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/gin-gonic/gin"
	"github.com/tochti/gin-angular-kauth"
	"github.com/tochti/smem"
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

	sessionStore := smem.NewStore()
	expires := time.Now().Add(1 * time.Hour)
	tmp := strconv.FormatInt(userID, 10)
	session, err := sessionStore.NewSession(tmp, expires)
	if err != nil {
		t.Fatal(err)
	}

	srv := gin.New()
	signedIn := kauth.SignedIn(&sessionStore)
	h := NewAppHandler(app, AppendSeriesListHandler)
	srv.POST("/", signedIn(h))

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
	resp := req.SendWithToken("POST", "/", session.Token())

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	expect := NewSuccessResponse("")
	err = EqualResponse(expect, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ReadSeriesListHandler_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)
	expect := SeriesList{
		{0, "Mr. Robot", "robot.png"},
		{1, "Narcos", "narcos.png"},
	}

	m := `SELECT series.ID as ID, series.Title as Title, series.Image as Image FROM %v as series, %v as list`
	q := fmt.Sprintf(m, SeriesTable, SeriesListTable)
	rows := sqlmock.NewRows([]string{"ID", "Title", "Image"})

	for _, s := range expect {
		rows.AddRow(s.ID, s.Title, s.Image)
	}
	mock.ExpectQuery(q).WillReturnRows(rows)

	app := AppCtx{
		DB: db,
	}

	sessionStore := smem.NewStore()
	expires := time.Now().Add(1 * time.Hour)
	tmp := strconv.FormatInt(userID, 10)
	session, err := sessionStore.NewSession(tmp, expires)
	if err != nil {
		t.Fatal(err)
	}

	srv := gin.New()
	readSeriesHandler := NewAppHandler(app, ReadSeriesListHandler)
	signedIn := kauth.SignedIn(&sessionStore)
	h := signedIn(readSeriesHandler)
	srv.GET("/", h)

	body := `
	{
		"Data": [
			{
				"ID": 0,
				"Title": "Mr. Robot", 
				"Image": "robot.png"
			},
			{
				"ID": 1,
				"Title": "Narcos",
				"Image": "narcos.png"
			}
		]
	}
	`

	req := TestRequest{
		Body:    body,
		Handler: srv,
		Header:  http.Header{},
	}
	resp := req.SendWithToken("GET", "/", session.Token())

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	expectResp := NewSuccessResponse(expect)
	err = EqualResponse(expectResp, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_POST_UpdateLastWatched_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)
	seriesID := int64(2)
	lastSession := 3
	lastEpisode := 4

	q := fmt.Sprintf("REPLACE INTO %v", LastWatchedTable)
	mock.ExpectExec(q).
		WithArgs(userID, seriesID, lastSession, lastEpisode).
		WillReturnResult(sqlmock.NewResult(0, 1))

	sessionStore := smem.NewStore()
	expires := time.Now().Add(1 * time.Hour)
	tmp := strconv.FormatInt(userID, 10)
	session, err := sessionStore.NewSession(tmp, expires)
	if err != nil {
		t.Fatal(err)
	}

	app := AppCtx{
		DB: db,
	}
	srv := gin.New()
	signedIn := kauth.SignedIn(&sessionStore)
	lastWatchedHandler := NewAppHandler(app, UpdateLastWatchedHandler)
	srv.POST("/", signedIn(lastWatchedHandler))

	body := `
	{
		"Data": {
			"SeriesID": 2,
			"Session": 3,
			"Episode": 4
		}
	}
	`

	req := TestRequest{
		Body:    body,
		Handler: srv,
		Header:  http.Header{},
	}
	resp := req.SendWithToken("POST", "/", session.Token())

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	expect := NewSuccessResponse(LastWatched{
		UserID:   userID,
		SeriesID: seriesID,
		Session:  lastSession,
		Episode:  lastEpisode,
	})
	err = EqualResponse(expect, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_GET_LastWatchedList_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)

	expect := LastWatchedList{
		{userID, int64(1), 2, 3},
		{userID, int64(2), 4, 5},
	}

	s := "SELECT Series_ID, Session, Episode FROM %v"
	q := fmt.Sprintf(s, LastWatchedTable)
	rows := sqlmock.NewRows([]string{
		"Series_ID", "Session", "Episode",
	})

	for _, s := range expect {
		rows.AddRow(s.SeriesID, s.Session, s.Episode)
	}

	mock.ExpectQuery(q).WillReturnRows(rows)
	sessionStore := smem.NewStore()
	expires := time.Now().Add(1 * time.Hour)
	tmp := strconv.FormatInt(userID, 10)
	session, err := sessionStore.NewSession(tmp, expires)
	if err != nil {
		t.Fatal(err)
	}

	app := AppCtx{
		DB: db,
	}

	srv := gin.New()
	signedIn := kauth.SignedIn(&sessionStore)
	h := NewAppHandler(app, LastWatchedListHandler)
	srv.GET("/", signedIn(h))

	req := TestRequest{
		Body:    "",
		Handler: srv,
		Header:  http.Header{},
	}
	resp := req.SendWithToken("GET", "/", session.Token())

	if 200 != resp.Code {
		t.Fatal("Expect 200 was", resp.Code)
	}

	expectResp := NewSuccessResponse(expect)
	err = EqualResponse(expectResp, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}
