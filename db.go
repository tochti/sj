package sj

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/tochti/gin-angular-kauth"
)

const (
	SeriesTable           = "Series"
	EpisodesResourceTable = "EpisodeResource"
	UserTable             = "User"
)

type (
	Series struct {
		ID    int64
		Title string
		Image string
	}

	EpisodeResource struct {
		ID       int64
		SeriesID int64
		Name     string
		URL      string
	}

	Episode struct {
		ID       int64
		SeriesID int64
		Title    string
		Episode  int
		Session  int
	}

	User struct {
		ID       int64
		Name     string
		Password string
	}

	kauthUser struct {
		id       string
		password string
	}

	userStore struct {
		db   *sql.DB
		user User
	}
)

func NewSeries(db *sql.DB, s Series) (int64, error) {
	err := db.Ping()
	if err != nil {
		return -1, err
	}

	q := fmt.Sprintf("INSERT INTO %v VALUES(?, ?)", SeriesTable)
	res, err := db.Exec(q, s.Title, s.Image)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func ReadSeries(db *sql.DB, id int64) (Series, error) {
	err := db.Ping()
	if err != nil {
		return Series{}, err
	}

	var title string
	var image string
	q := fmt.Sprintf("SELECT Title, Image FROM %v WHERE ID = %v", SeriesTable, id)
	err = db.QueryRow(q).Scan(&title, &image)
	if err != nil {
		return Series{}, err
	}

	s := Series{
		ID:    id,
		Title: title,
		Image: image,
	}

	return s, nil
}

func FindSeriesByTitle(db *sql.DB, t string) (Series, error) {

	var id int64
	var title string
	var image string

	q := fmt.Sprintf("SELECT ID, Title, Image FROM %v WHERE Title = '%v'", SeriesTable, t)
	err := db.QueryRow(q).Scan(&id, &title, &image)
	if err != nil {
		return Series{}, err
	}

	s := Series{
		ID:    id,
		Title: title,
		Image: image,
	}

	return s, nil
}

func NewEpisodeResource(db *sql.DB, r EpisodeResource) (int64, error) {
	err := db.Ping()
	if err != nil {
		return -1, err
	}

	q := fmt.Sprintf("INSERT INTO %v VALUES(?, ?)", EpisodesResourceTable)
	rsrc, err := db.Exec(q, r.Name, r.URL)
	if err != nil {
		return -1, err
	}

	id, err := rsrc.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil

}

func ReadEpisodeResource(db *sql.DB, id int64) (EpisodeResource, error) {
	err := db.Ping()
	if err != nil {
		return EpisodeResource{}, err
	}

	m := "SELECT Series_ID, Name, URL FROM %v WHERE ID = %v"
	q := fmt.Sprintf(m, EpisodesResourceTable, id)
	var seriesID int64
	var name string
	var url string
	err = db.QueryRow(q).Scan(&seriesID, &name, &url)
	if err != nil {
		return EpisodeResource{}, err
	}

	r := EpisodeResource{
		ID:       id,
		SeriesID: seriesID,
		Name:     name,
		URL:      url,
	}

	return r, nil

}

func NewUser(db *sql.DB, user User) (int64, error) {
	err := db.Ping()
	if err != nil {
		return -1, err
	}

	m := "INSERT INTO %v VALUES (?,?)"
	q := fmt.Sprintf(m, UserTable)
	pass := NewSha512Password(user.Password)
	rsrc, err := db.Exec(q, user.Name, pass)
	if err != nil {
		return -1, err
	}

	id, err := rsrc.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func ReadUser(db *sql.DB, id int64) (User, error) {
	err := db.Ping()
	if err != nil {
		return User{}, err
	}

	m := "SELECT ID,Name,Password FROM %v WHERE ID=%v"
	q := fmt.Sprintf(m, UserTable, id)

	var idTmp int64
	var name string
	var pass string

	err = db.QueryRow(q).Scan(&idTmp, &name, &pass)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       idTmp,
		Name:     name,
		Password: pass,
	}

	return user, nil
}

func FindUserByName(db *sql.DB, name string) (User, error) {
	err := db.Ping()
	if err != nil {
		return User{}, err
	}

	m := "SELECT ID,Name,Password FROM %v WHERE Name=%v"
	q := fmt.Sprintf(m, UserTable, name)

	var id int64
	var nameTmp string
	var pass string

	err = db.QueryRow(q).Scan(&id, &nameTmp, &pass)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       id,
		Name:     nameTmp,
		Password: pass,
	}

	return user, nil
}

func NewUserStore(db *sql.DB) kauth.UserStore {
	return &userStore{
		db: db,
	}
}

func (s *userStore) FindUser(name string) (kauth.User, error) {
	user, err := FindUserByName(s.db, name)
	if err != nil {
		return nil, err
	}

	s.user = user

	kuser := kauthUser{
		id:       strconv.FormatInt(user.ID, 10),
		password: user.Password,
	}

	return kuser, nil
}

func (s *userStore) ValidPassword(pass string) bool {
	return s.user.Password == NewSha512Password(pass)
}

func (u kauthUser) ID() string {
	return u.id
}

func (u kauthUser) Password() string {
	return u.password
}
