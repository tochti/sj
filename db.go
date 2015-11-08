package sj

import (
	"database/sql"
	"fmt"
)

const (
	SeriesTable           = "Series"
	EpisodesResourceTable = "EpisodeResource"
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
	query := fmt.Sprintf("SELECT Title, Image FROM %v WHERE ID = %v", SeriesTable, id)
	err = db.QueryRow(query).Scan(&title, &image)
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

	query := fmt.Sprintf("INSERT INTO %v VALUES(?, ?)", EpisodesResourceTable)
	rsrc, err := db.Exec(query, r.Name, r.URL)
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
	query := fmt.Sprintf(m, EpisodesResourceTable, id)
	var seriesID int64
	var name string
	var url string
	err = db.QueryRow(query).Scan(&seriesID, &name, &url)
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
