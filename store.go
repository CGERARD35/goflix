package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Store interface {
	Open() error
	Close() error

	GetMovies() ([]*Movie, error)
	GetMovieById(id int64) (*Movie, error)
	CreateMovie(m *Movie) error
	DeleteMovieById(id int64) error
	UpdateMovie(id int64, m Movie) error

	FindUser(username string, password string) (bool, error)
}

type dbStore struct {
	db *sqlx.DB
}

var schema = `
CREATE TABLE IF NOT EXISTS movie
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT,
	release_date TEXT,
	duration INTEGER,
	trailer_url TEXT
);

CREATE TABLE IF NOT EXISTS user
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user TEXT,
	password TEXT
);
`

func (store *dbStore) Open() error {
	db, err := sqlx.Connect("sqlite3", "goflix.db")
	if err != nil {
		return err
	}
	log.Println("Connect to DB")
	db.MustExec(schema)
	store.db = db
	return nil
}

func (store *dbStore) Close() error {
	return store.db.Close()
}

func (store *dbStore) GetMovies() ([]*Movie, error) {
	var movies []*Movie
	err := store.db.Select(&movies, "SELECT * FROM movie")
	if err != nil {
		return movies, err
	}
	return movies, nil
}

func (store *dbStore) GetMovieById(id int64) (*Movie, error) {
	var movie = &Movie{}
	err := store.db.Get(movie, "SELECT * FROM movie WHERE id=$1", id)
	if err != nil {
		return movie, nil
	}
	return movie, nil
}

func (store *dbStore) FindMovieById(id int64) (*Movie, error) {
	query := `SELECT id, title, release_date, duration, trailer_url FROM movies WHERE id = ?`
	var movie Movie
	row := store.db.QueryRow(query, id)
	err := row.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &movie.Duration, &movie.TrailerURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("aucun film trouvé avec l'ID %d", id)
		}
		return nil, err
	}
	return &movie, nil
}

func (store *dbStore) UpdateMovie(id int64, m Movie) error {
	stmt, err := store.db.Prepare(`UPDATE movie SET title = ?, release_date = ?, duration = ?, trailer_url = ?
		WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.Title, m.ReleaseDate, m.Duration, m.TrailerURL, id)
	return err
}

func (store *dbStore) CreateMovie(m *Movie) error {
	res, err := store.db.Exec("INSERT INTO movie (title, release_date, duration, trailer_url) VALUES (?, ?, ?, ?)",
		m.Title, m.ReleaseDate, m.Duration, m.TrailerURL)
	if err != nil {
		return err
	}

	m.ID, err = res.LastInsertId()
	return err
}

func (store *dbStore) DeleteMovieById(id int64) error {
	movie := &Movie{}
	err := store.db.Get(movie, "SELECT * FROM movie WHERE id = $1", id)
	if err != nil {
		return err
	}
	_, err = store.db.Exec("DELETE FROM movie WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (store *dbStore) FindUser(username string, password string) (bool, error) {
	var count int
	err := store.db.Get(&count, "SELECT COUNT(id) FROM user WHERE user=$1 AND password=$2", username, password)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}
