package sj

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

type (
	Specs struct {
		Host      string
		Port      int
		PublicDir string `envconfig:"public_dir"`
		ImageDir  string `envconfig:"image_dir"`
		DBHost    string `envconfig:"db_host"`
		DBPort    int    `envconfig:"db_port"`
		DBUser    string `envconfig:"db_user"`
		DBPass    string `envconfig:"db_pass"`
		DBName    string `envconfig:"db_name"`
	}

	AppCtx struct {
		Specs Specs
		DB    *sql.DB
	}

	JSONRequest struct {
		Data interface{}
	}

	SeriesListRequestData struct {
		SeriesID int64
	}
)

func NewApp(name string) (AppCtx, error) {
	specs := Specs{}
	err := envconfig.Process(name, &specs)
	if err != nil {
		return AppCtx{}, err
	}

	url := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		specs.DBUser,
		specs.DBPass,
		specs.DBHost,
		specs.DBPort,
		specs.DBName,
	)

	db, err := sql.Open("mysql", url)
	if err != nil {
		return AppCtx{}, err
	}

	ctx := AppCtx{
		Specs: specs,
		DB:    db,
	}

	return ctx, nil
}

func SaveImage(url, p string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(resp.Body)
	buf := bytes.NewBuffer([]byte{})

	_, err = reader.WriteTo(buf)
	if err != nil {
		return "", err
	}

	content := buf.Bytes()

	hash := NewSha1Hash(content)
	ext := path.Ext(url)
	filename := hash + ext
	file := path.Join(p, filename)

	// If the image already exists we don't need to save it again
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return filename, nil
	}

	err = ioutil.WriteFile(file, content, 0755)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func removeImage(p string) error {
	return os.Remove(p)
}

func NewSha1Hash(by []byte) string {
	hash := sha1.Sum(by)
	hex := fmt.Sprintf("%x", hash)
	return hex
}

func ParseJSONRequest(r *http.Request) (JSONRequest, error) {
	buf := bytes.NewBuffer([]byte{})
	_, err := buf.ReadFrom(r.Body)

	req := JSONRequest{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		return JSONRequest{}, err
	}

	return req, nil
}

func ExistsFields(s map[string]interface{}, f []string) error {
	for _, v := range f {
		_, ok := s[v]
		if !ok {
			return NewMissingFieldError(v)
		}
	}

	return nil
}

func NewMissingFieldError(field string) error {
	msg := fmt.Sprintf("%v is missing", field)
	return errors.New(msg)
}

func ParseNewSeriesRequest(c *gin.Context) (Series, error) {
	req, err := ParseJSONRequest(c.Request)
	if err != nil {
		return Series{}, err
	}

	tmp, ok := req.Data.(map[string]interface{})

	err = ExistsFields(tmp, []string{"Title", "Image"})
	if err != nil {
		return Series{}, err
	}

	title, ok := tmp["Title"].(string)
	if !ok {
		m := "Wrong value in Title"
		return Series{}, errors.New(m)
	}

	image, ok := tmp["Image"].(string)
	if !ok {
		m := "Wrong value in Image"
		return Series{}, errors.New(m)
	}

	s := Series{
		Title: title,
		Image: image,
	}

	return s, nil
}

func ParseNewUserRequest(c *gin.Context) (User, error) {
	req, err := ParseJSONRequest(c.Request)
	if err != nil {
		return User{}, err
	}

	tmp, ok := req.Data.(map[string]interface{})
	err = ExistsFields(tmp, []string{"Name", "Password"})
	if err != nil {
		return User{}, err
	}

	name, ok := tmp["Name"].(string)
	if !ok {
		m := "Wrong value in Name"
		return User{}, errors.New(m)
	}

	pass, ok := tmp["Password"].(string)
	if !ok {
		m := "Wrong value in Password"
		return User{}, errors.New(m)
	}

	u := User{
		Name:     name,
		Password: pass,
	}

	return u, nil
}

func ParseAppendSeriesListRequest(c *gin.Context) (SeriesListRequestData, error) {
	req, err := ParseJSONRequest(c.Request)
	if err != nil {
		return SeriesListRequestData{}, err
	}

	tmp, ok := req.Data.(map[string]interface{})
	err = ExistsFields(tmp, []string{"SeriesID"})
	if err != nil {
		return SeriesListRequestData{}, err
	}

	seriesID, ok := tmp["SeriesID"].(float64)
	if !ok {
		m := "Wrong value in UserID"
		return SeriesListRequestData{}, errors.New(m)
	}

	s := SeriesListRequestData{
		SeriesID: int64(seriesID),
	}

	return s, nil
}

func ParseUpdateLastWatchedRequest(c *gin.Context) (LastWatched, error) {
	req, err := ParseJSONRequest(c.Request)
	if err != nil {
		return LastWatched{}, err
	}

	tmp, ok := req.Data.(map[string]interface{})
	err = ExistsFields(tmp, []string{
		"SeriesID", "LastSession", "LastEpisode",
	})
	if err != nil {
		return LastWatched{}, err
	}

	seriesID, ok := tmp["SeriesID"].(float64)
	if !ok {
		m := "Wrong value in SeriesID"
		return LastWatched{}, errors.New(m)
	}

	lastSession, ok := tmp["LastSession"].(float64)
	if !ok {
		m := "Wrong value in LastSession"
		return LastWatched{}, errors.New(m)
	}

	lastEpisode, ok := tmp["LastEpisode"].(float64)
	if !ok {
		m := "Wrong value in LastEpisode"
		return LastWatched{}, errors.New(m)
	}

	w := LastWatched{
		SeriesID:    int64(seriesID),
		LastSession: int(lastSession),
		LastEpisode: int(lastEpisode),
	}

	return w, nil
}
func NewSha512Password(pass string) string {
	hash := sha512.New()
	tmp := hash.Sum([]byte(pass))
	passHash := fmt.Sprintf("%x", tmp)
	return passHash
}
