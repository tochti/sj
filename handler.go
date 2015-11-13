package sj

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tochti/gin-angular-kauth"
)

type (
	SuccessResponse struct {
		Status string
		Data   interface{}
	}

	FailResponse struct {
		Status string
		Err    string
	}

	AppHandler func(AppCtx, *gin.Context) error
)

func NewSuccessResponse(data interface{}) SuccessResponse {
	resp := SuccessResponse{
		Status: "success",
		Data:   data,
	}

	return resp
}

func NewFailResponse(err error) FailResponse {
	resp := FailResponse{
		Status: "fail",
		Err:    err.Error(),
	}

	return resp
}

func NewAppHandler(app AppCtx, fn AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := fn(app, c)
		if err != nil {
			resp := NewFailResponse(err)
			c.JSON(http.StatusOK, resp)
		}
	}
}

func NewSeriesHandler(app AppCtx, c *gin.Context) error {
	s, err := ParseNewSeriesRequest(c)
	if err != nil {
		return err
	}

	id, err := NewSeries(app.DB, s)
	if err != nil {
		return err
	}

	s.ID = id

	err = SaveImage(s.Image, app.Specs.ImageDir)
	if err != nil {
		return err
	}

	resp := NewSuccessResponse(s)
	c.JSON(http.StatusOK, resp)

	return nil
}

func ReadSeriesHandler(app AppCtx, c *gin.Context) error {

	tmp := c.Params.ByName("id")

	id, err := strconv.Atoi(tmp)
	if err != nil {
		return err
	}

	s, err := ReadSeries(app.DB, int64(id))
	if err != nil {
		return err
	}

	resp := NewSuccessResponse(s)
	c.JSON(http.StatusOK, resp)

	return nil
}

func NewUserHandler(app AppCtx, c *gin.Context) error {
	user, err := ParseNewUserRequest(c)
	if err != nil {
		return err
	}

	_, err = FindUserByName(app.DB, user.Name)
	if err == nil || err != sql.ErrNoRows {
		if err != nil {
			return err
		}

		m := fmt.Sprintf("User %v already exists", user.Name)
		return errors.New(m)
	}

	id, err := NewUser(app.DB, user)
	if err != nil {
		return err
	}

	u := User{
		ID:   id,
		Name: user.Name,
	}

	resp := NewSuccessResponse(u)
	c.JSON(http.StatusOK, resp)

	return nil

}

func AppendSeriesListHandler(app AppCtx, c *gin.Context) error {
	data, err := ParseAppendSeriesListRequest(c)
	if err != nil {
		return err
	}

	err = AppendSeriesList(app.DB, data.UserID, data.SeriesID)
	if err != nil {
		return err
	}

	resp := NewSuccessResponse("")
	c.JSON(http.StatusOK, resp)

	return nil
}

func ReadSeriesListHandler(app AppCtx, c *gin.Context) error {

	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}
	id, err := strconv.Atoi(session.UserID())
	if err != nil {
		return err
	}

	sList, err := ReadSeriesList(app.DB, int64(id))
	if err != nil {
		return err
	}

	resp := NewSuccessResponse(sList)
	c.JSON(http.StatusOK, resp)

	return nil
}
