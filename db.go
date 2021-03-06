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
	SeriesListTable       = "SeriesList"
	LastWatchedTable      = "LastWatched"
)

type (
	Series struct {
		ID    int64
		Title string
		Image string
	}

	SeriesList []Series

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

	LastWatched struct {
		UserID   int64
		SeriesID int64
		Session  int
		Episode  int
	}

	LastWatchedList []LastWatched
)

func NewSeries(db *sql.DB, s Series) (int64, error) {
	err := db.Ping()
	if err != nil {
		return -1, err
	}

	q := fmt.Sprintf("INSERT INTO %v (Title,Image) VALUES(?, ?)", SeriesTable)
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
	q := fmt.Sprintf("SELECT Title, Image FROM %v WHERE ID = ?", SeriesTable)
	err = db.QueryRow(q, id).Scan(&title, &image)
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

func RemoveSeries(db *sql.DB, id int64) error {
	if err := db.Ping(); err != nil {
		return err
	}

	s := "DELETE FROM %v WHERE ID = ?"
	q := fmt.Sprintf(s, SeriesTable)
	if _, err := db.Exec(q, id); err != nil {
		return err
	}

	return nil
}

func FindSeriesByTitle(db *sql.DB, t string) (Series, error) {

	var id int64
	var title string
	var image string

	q := fmt.Sprintf("SELECT ID, Title, Image FROM %v WHERE Title = ?", SeriesTable)
	err := db.QueryRow(q, t).Scan(&id, &title, &image)
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

	m := "SELECT Series_ID, Name, URL FROM %v WHERE ID = ?"
	q := fmt.Sprintf(m, EpisodesResourceTable)
	var seriesID int64
	var name string
	var url string
	err = db.QueryRow(q, id).Scan(&seriesID, &name, &url)
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

	m := "INSERT INTO %v (Name,Password) VALUES (?,?)"
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

	m := "SELECT ID,Name,Password FROM %v WHERE ID = ?"
	q := fmt.Sprintf(m, UserTable)

	var idTmp int64
	var name string
	var pass string

	err = db.QueryRow(q, id).Scan(&idTmp, &name, &pass)
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

	m := "SELECT ID,Name,Password FROM %v WHERE Name = ?"
	q := fmt.Sprintf(m, UserTable)

	var id int64
	var nameTmp string
	var pass string

	err = db.QueryRow(q, name).Scan(&id, &nameTmp, &pass)
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

func (u kauthUser) ValidPassword(pass string) bool {
	return u.password == NewSha512Password(pass)
}

func (u kauthUser) ID() string {
	return u.id
}

func (u kauthUser) Password() string {
	return u.password
}

func AppendSeriesList(db *sql.DB, userID, seriesID int64) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	q := fmt.Sprintf("INSERT INTO %v VALUES(?, ?)", SeriesListTable)
	_, err = db.Exec(q, userID, seriesID)
	if err != nil {
		return err
	}

	return nil
}

func RemoveSeriesList(db *sql.DB, userID, seriesID int64) (int64, error) {
	err := db.Ping()
	if err != nil {
		return 0, err
	}

	s := "DELETE FROM %v WHERE User_ID = ? AND Series_ID = ?"
	q := fmt.Sprintf(s, SeriesListTable)
	rsrc, err := db.Exec(q, userID, seriesID)
	if err != nil {
		return 0, err
	}

	c, err := rsrc.RowsAffected()
	if err != nil {
		return 0, err
	}

	return c, nil
}

func ReadSeriesList(db *sql.DB, userID int64) (SeriesList, error) {
	m := `
	SELECT series.ID as ID, series.Title as Title, series.Image as Image
	FROM %v as series, %v as list 
	WHERE list.User_ID = ? 
	AND series.ID=list.Series_ID
	`
	q := fmt.Sprintf(m, SeriesTable, SeriesListTable)

	rows, err := db.Query(q, userID)
	if err != nil {
		return SeriesList{}, err
	}
	defer rows.Close()

	sList := SeriesList{}
	for rows.Next() {
		var id int64
		var title string
		var image string
		err := rows.Scan(&id, &title, &image)
		if err != nil {
			return SeriesList{}, err
		}

		series := Series{
			ID:    id,
			Title: title,
			Image: image,
		}
		sList = append(sList, series)
	}

	return sList, nil
}

func UpdateLastWatched(db *sql.DB, lastWatched LastWatched) error {

	err := db.Ping()
	if err != nil {
		return err
	}

	s := "REPLACE INTO %v VALUES (?, ?, ?, ?)"
	q := fmt.Sprintf(s, LastWatchedTable)
	_, err = db.Exec(q, lastWatched.UserID, lastWatched.SeriesID,
		lastWatched.Session, lastWatched.Episode)
	if err != nil {
		return err
	}

	return nil
}

func ReadLastWatchedList(db *sql.DB, userID int64) (LastWatchedList, error) {
	err := db.Ping()
	if err != nil {
		return LastWatchedList{}, err
	}

	s := `
	SELECT Series_ID, Session, Episode
	FROM %v
	WHERE User_ID = ? 
	`
	q := fmt.Sprintf(s, LastWatchedTable)
	rows, err := db.Query(q, userID)
	if err != nil {
		return LastWatchedList{}, err
	}
	defer rows.Close()

	wList := LastWatchedList{}
	for rows.Next() {
		var seriesID int64
		var lastSession int
		var lastEpisode int

		err := rows.Scan(&seriesID, &lastSession, &lastEpisode)
		if err != nil {
			return LastWatchedList{}, err
		}

		tmp := LastWatched{
			UserID:   userID,
			SeriesID: seriesID,
			Session:  lastSession,
			Episode:  lastEpisode,
		}
		wList = append(wList, tmp)

	}

	return wList, nil
}

func CountSeriesWithImage(db *sql.DB, image string) (int, error) {
	if err := db.Ping(); err != nil {
		return 0, err
	}

	s := "SELECT COUNT(ID) as Images FROM %v WHERE Image = ?"
	q := fmt.Sprintf(s, SeriesTable)
	var amount int
	err := db.QueryRow(q, image).Scan(&amount)
	if err != nil {
		return 0, err
	}

	return amount, nil

}
