package sj

import (
	"bytes"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
)

func NewApp(name string) (AppCtx, error) {
	specs := Specs{}
	err := envconfig.Process(name, &specs)
	if err != nil {
		return AppCtx{}, err
	}

	url := fmt.Sprintf("%v:%v@%v:%v/%v",
		specs.DBUser,
		specs.DBPass,
		specs.DBHost,
		specs.DBPort,
		specs.DBName,
	)
	db, err := sql.Open("mysql", url)

	ctx := AppCtx{
		Specs: specs,
		DB:    db,
	}

	return ctx, nil
}

func SaveImage(url, p string) error {
	return nil
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

func NewSha512Password(pass string) string {
	hash := sha512.New()
	tmp := hash.Sum([]byte(pass))
	passHash := fmt.Sprintf("%x", tmp)
	return passHash
}
