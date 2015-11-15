package sj

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path"
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

	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}

	s, err := ParseNewSeriesRequest(c)
	if err != nil {
		return err
	}

	imgDir := app.Specs.ImageDir
	name, err := SaveImage(s.Image, imgDir)
	if err != nil {
		return err
	}
	imgPath := path.Join(imgDir, name)

	s.Image = name

	seriesID, err := NewSeries(app.DB, s)
	if err != nil {
		removeImage(imgPath)
		return err
	}

	userID, err := strconv.Atoi(session.UserID())
	if err != nil {
		// todo(tochti):remove series
		removeImage(imgPath)
		return err
	}

	err = AppendSeriesList(app.DB, int64(userID), seriesID)
	if err != nil {
		// todo(tochti):remove series
		removeImage(imgPath)
		return err
	}

	s.ID = seriesID

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

func RemoveSeriesHandler(app AppCtx, c *gin.Context) error {
	idParam := c.Params.ByName("id")
	tmp, err := strconv.Atoi(idParam)
	if err != nil {
		return err
	}
	seriesID := int64(tmp)

	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}
	tmp, err = strconv.Atoi(session.UserID())
	if err != nil {
		return err
	}
	userID := int64(tmp)

	series, err := ReadSeries(app.DB, seriesID)
	if err != nil {
		return err
	}

	affected, err := RemoveSeriesList(app.DB, userID, seriesID)
	if err != nil {
		return err
	}

	if affected < 1 {
		return errors.New("Cannot found Series")
	}

	err = RemoveSeries(app.DB, seriesID)
	if err != nil {
		err2 := AppendSeriesList(app.DB, userID, seriesID)
		if err2 != nil {
			return err2
		}

		return err
	}

	count, err := CountSeriesWithImage(app.DB, series.Image)
	if err != nil {
		return err
	}

	if count > 1 {
		removeImage(path.Join(app.Specs.ImageDir, series.Image))
	}

	resp := NewSuccessResponse(series)
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
	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}
	userID, err := strconv.Atoi(session.UserID())
	if err != nil {
		return err
	}

	data, err := ParseAppendSeriesListRequest(c)
	if err != nil {
		return err
	}

	err = AppendSeriesList(app.DB, int64(userID), data.SeriesID)
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

func UpdateLastWatchedHandler(app AppCtx, c *gin.Context) error {
	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}
	userID, err := strconv.Atoi(session.UserID())
	if err != nil {
		return err
	}

	lastWatched, err := ParseUpdateLastWatchedRequest(c)
	if err != nil {
		return err
	}

	lastWatched.UserID = int64(userID)

	err = UpdateLastWatched(app.DB, lastWatched)
	if err != nil {
		return err
	}

	resp := NewSuccessResponse(lastWatched)
	c.JSON(http.StatusOK, resp)

	return nil
}

func LastWatchedListHandler(app AppCtx, c *gin.Context) error {
	session, err := kauth.ReadSession(c)
	if err != nil {
		return err
	}
	userID, err := strconv.Atoi(session.UserID())
	if err != nil {
		return err
	}

	watchedList, err := ReadLastWatchedList(app.DB, int64(userID))
	if err != nil {
		return err
	}

	resp := NewSuccessResponse(watchedList)
	c.JSON(http.StatusOK, resp)

	return nil
}
