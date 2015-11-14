package sj

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var (
	series = Series{
		ID:    1,
		Title: "Mr. Robot",
		Image: "http://photo/img.png",
	}

	resource = EpisodeResource{
		ID:       1,
		SeriesID: 1,
		Name:     "sejun",
		URL:      "http://sejun",
	}
)

func EqualSeries(s1, s2 Series) error {
	if s1.ID != s2.ID ||
		s1.Title != s2.Title ||
		s1.Image != s2.Image {
		m := fmt.Sprintf("Expect %v was %v", s1, s2)
		return errors.New(m)
	}

	return nil

}

func EqualEpisodeResource(r1, r2 EpisodeResource) error {
	if r1.ID != r2.ID ||
		r1.Name != r2.Name ||
		r1.URL != r2.URL {
		m := fmt.Sprintf("Expect %v was %v", r1, r2)
		return errors.New(m)
	}

	return nil
}

func EqualUser(u1, u2 User) error {
	if u1.ID != u2.ID ||
		u1.Name != u2.Name ||
		u1.Password != u2.Password {
		m := fmt.Sprintf("Expect %v was %v", u1, u2)
		return errors.New(m)
	}

	return nil
}

func EqualSeriesList(s1, s2 SeriesList) error {
	if len(s1) != len(s2) {
		e := fmt.Sprintf("Expect length %v was %v", len(s1), len(s2))
		errors.New(e)
	}

	for i, _ := range s1 {
		err := EqualSeries(s1[i], s2[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func Test_NewSeries_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO %v", SeriesTable)
	mock.ExpectExec(query).
		WithArgs(series.Title, series.Image).
		WillReturnResult(sqlmock.NewResult(series.ID, 1))

	s := Series{
		Title: series.Title,
		Image: series.Image,
	}

	id, err := NewSeries(db, s)
	if err != nil {
		t.Fatal(err)
	}

	if id != 1 {
		t.Fatal("Expect ID")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_ReadSeries_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT Title, Image FROM %v", SeriesTable)
	rows := sqlmock.NewRows([]string{"Title", "Image"}).
		AddRow(series.Title, series.Image)
	mock.ExpectQuery(query).WillReturnRows(rows)

	s, err := ReadSeries(db, series.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = EqualSeries(series, s)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_FindSeries_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT ID, Title, Image FROM %v", SeriesTable)
	rows := sqlmock.NewRows([]string{"ID", "Title", "Image"}).
		AddRow(series.ID, series.Title, series.Image)
	mock.ExpectQuery(query).WillReturnRows(rows)

	s, err := FindSeriesByTitle(db, series.Title)
	if err != nil {
		t.Fatal(err)
	}

	err = EqualSeries(series, s)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_NewEpisodeResource_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO %v", EpisodesResourceTable)
	mock.ExpectExec(query).
		WithArgs(resource.Name, resource.URL).
		WillReturnResult(sqlmock.NewResult(resource.ID, 1))

	rsrc := EpisodeResource{
		Name: resource.Name,
		URL:  resource.URL,
	}
	id, err := NewEpisodeResource(db, rsrc)

	if id != 1 {
		t.Fatal("Expect ID")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func Test_ReadEpisodeResource_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT Series_ID, Name, URL FROM %v", EpisodesResourceTable)
	rows := sqlmock.NewRows([]string{"Series_ID", "Name", "URL"}).
		AddRow(resource.SeriesID, resource.Name, resource.URL)
	mock.ExpectQuery(query).WillReturnRows(rows)

	r, err := ReadEpisodeResource(db, resource.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := EqualEpisodeResource(resource, r); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func Test_NewUser_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user := User{
		ID:       12,
		Name:     "test",
		Password: "test",
	}

	query := fmt.Sprintf("INSERT INTO %v", UserTable)
	mock.ExpectExec(query).
		WithArgs(user.Name, NewSha512Password(user.Password)).
		WillReturnResult(sqlmock.NewResult(user.ID, 1))

	id, err := NewUser(db, user)

	if id != user.ID {
		t.Fatal("Expect", user.ID, "was", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_ReadUser_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user := User{
		ID:       14,
		Name:     "peacemaker",
		Password: "Fuckoff",
	}

	m := "SELECT ID,Name,Password FROM %v"
	q := fmt.Sprintf(m, UserTable)
	rows := sqlmock.NewRows([]string{"ID", "Name", "Password"}).
		AddRow(user.ID, user.Name, user.Password)
	mock.ExpectQuery(q).WillReturnRows(rows)

	result, err := ReadUser(db, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = EqualUser(user, result)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}

func Test_FindUserByName_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user := User{
		ID:       14,
		Name:     "peacemaker",
		Password: "Fuckoff",
	}

	m := "SELECT ID,Name,Password FROM %v"
	q := fmt.Sprintf(m, UserTable)
	rows := sqlmock.NewRows([]string{"ID", "Name", "Password"}).
		AddRow(user.ID, user.Name, user.Password)
	mock.ExpectQuery(q).WillReturnRows(rows)

	result, err := FindUserByName(db, user.Name)
	if err != nil {
		t.Fatal(err)
	}

	err = EqualUser(user, result)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_UserStoreFindUserAndValidatedPassword_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user := User{
		ID:       14,
		Name:     "peacemaker",
		Password: "Fuckoff",
	}

	m := "SELECT ID,Name,Password FROM %v"
	q := fmt.Sprintf(m, UserTable)
	rows := sqlmock.NewRows([]string{"ID", "Name", "Password"}).
		AddRow(user.ID, user.Name, NewSha512Password(user.Password))
	mock.ExpectQuery(q).WillReturnRows(rows)

	userStore := NewUserStore(db)

	result, err := userStore.FindUser(user.Name)
	if err != nil {
		t.Fatal(err)
	}

	if strconv.FormatInt(user.ID, 10) != result.ID() ||
		NewSha512Password(user.Password) != result.Password() {
		t.Fatal("Expect", user, "was", result)
	}

	if !result.ValidPassword(user.Password) {
		t.Fatal("Expect", user.Password, "to be correct")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_AppendSeriesList_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	userID := int64(2)
	seriesID := int64(1)

	q := fmt.Sprintf("INSERT INTO %v", SeriesListTable)
	mock.ExpectExec(q).
		WithArgs(userID, seriesID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = AppendSeriesList(db, userID, seriesID)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func Test_ReadSeriesList_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)
	expect := SeriesList{
		{0, "Mr. Robot", "robot.png"},
		{1, "Narcos", "narcos.png"},
	}
	m := `SELECT series.ID as ID, series.Title as Title, series.Image as Image FROM %v as series, %v as list WHERE list.User_ID=%v AND series.ID=list.Series_ID`
	q := fmt.Sprintf(m, SeriesTable, SeriesListTable, userID)
	rows := sqlmock.NewRows([]string{"ID", "Title", "Image"})

	for _, s := range expect {
		rows.AddRow(s.ID, s.Title, s.Image)
	}

	mock.ExpectQuery(q).WillReturnRows(rows)

	seriesList, err := ReadSeriesList(db, userID)
	if err != nil {
		t.Fatal(err)
	}

	err = EqualSeriesList(expect, seriesList)
	if err != nil {
		t.Fatal(err)
	}

}

func Test_UpdateLastWatched_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	userID := int64(1)
	seriesID := int64(2)
	lastSession := 3
	lastEpisode := 4
	s := "REPLACE INTO %v"
	q := fmt.Sprintf(s, LastWatchedTable)
	mock.ExpectExec(q).
		WithArgs(userID, seriesID, lastSession, lastEpisode).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = UpdateLastWatched(db, userID, seriesID, lastSession, lastEpisode)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

}
