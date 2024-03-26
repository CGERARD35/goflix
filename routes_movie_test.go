package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStore struct {
	movieId int64
	movies  []*Movie
}

func (t TestStore) Open() error {
	return nil
}

func (t TestStore) Close() error {
	return nil
}

func (t TestStore) GetMovies() ([]*Movie, error) {
	return t.movies, nil
}

func (t TestStore) GetMovieById(id int64) (*Movie, error) {
	for _, m := range t.movies {
		if m.ID == id {
			return m, nil
		}
	}
	return nil, nil
}

func (t TestStore) CreateMovie(m *Movie) error {
	t.movieId++
	m.ID = t.movieId
	t.movies = append(t.movies, m)
	return nil
}

func (t *TestStore) UpdateMovie(id int64, m Movie) error {
	for i, movie := range t.movies {
		if movie.ID == m.ID {
			t.movies[i].Title = m.Title
			t.movies[i].ReleaseDate = m.ReleaseDate
			t.movies[i].Duration = m.Duration
			t.movies[i].TrailerURL = m.TrailerURL
			return nil
		}
	}
	return fmt.Errorf("film avec l'ID %d non trouvé", m.ID)
}

func (t *TestStore) FindMovieById(id int64) (*Movie, error) {
	for _, movie := range t.movies {
		if movie != nil && movie.ID == id {
			return movie, nil
		}
	}
	return nil, fmt.Errorf("film avec l'ID %d non trouvé", id)
}

func (t *TestStore) DeleteMovieById(id int64) error {
	found := false
	for i, m := range t.movies {
		if m.ID == id {
			t.movies = append(t.movies[:i], t.movies[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("film non trouvé")
	}

	return nil
}

func (t *TestStore) FindUser(username string, password string) (bool, error) {
	return true, nil
}

func TestFindMovieById(t *testing.T) {
	// Création d'un TestStore avec un film de test
	testStore := &TestStore{
		movies: []*Movie{
			{ID: 1, Title: "Test Movie", ReleaseDate: "2021-01-01", Duration: 120, TrailerURL: "http://example.com"},
		},
	}

	// Test de la récupération du film par son ID
	movie, err := testStore.FindMovieById(1)
	assert.NoError(t, err)
	assert.NotNil(t, movie)
	assert.Equal(t, int64(1), movie.ID)
	assert.Equal(t, "Test Movie", movie.Title)

	// Test de la récupération d'un film non existant par son ID
	_, err = testStore.FindMovieById(2)
	assert.Error(t, err)
}

func TestMovieCreateUnit(t *testing.T) {
	srv := newServer()
	srv.store = &TestStore{}

	p := struct {
		Title       string `json:"title"`
		ReleaseDate string `json:"release_date"`
		Duration    int    `json:"duration"`
		TrailerURL  string `json:"trailer_url"`
	}{
		Title:       "Inception",
		ReleaseDate: "2010-07-18",
		Duration:    148,
		TrailerURL:  "http://url",
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(p)
	assert.Nil(t, err)

	r := httptest.NewRequest("POST", "/api/movies/", &buf)
	w := httptest.NewRecorder()

	srv.handleMovieCreate()(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerUpdateMovie(t *testing.T) {
	testStore := &TestStore{
		movies: []*Movie{
			{ID: 1, Title: "Original Title", ReleaseDate: "2020-01-01", Duration: 120, TrailerURL: "http://example.com"},
		},
	}
	srv := newServer()
	srv.store = testStore

	movieUpdate := Movie{
		Title:       "Updated Title",
		ReleaseDate: "2020-02-02",
		Duration:    125,
		TrailerURL:  "http://example.com/updated",
	}
	body, err := json.Marshal(movieUpdate)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/movies/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r := httptest.NewRecorder()

	handler := srv.handleUpdateMovie()
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "Le statut HTTP attendu est 200 OK")

	updatedMovie, _ := testStore.FindMovieById(1)
	assert.NotNil(t, updatedMovie)
	assert.Equal(t, "Updated Title", updatedMovie.Title)
	assert.Equal(t, "2020-02-02", updatedMovie.ReleaseDate)
	assert.Equal(t, 125, updatedMovie.Duration)
	assert.Equal(t, "http://example.com/updated", updatedMovie.TrailerURL)
}

func TestHandleDeleteMovieById(t *testing.T) {
	testStore := &TestStore{
		movies: []*Movie{
			{ID: 1, Title: "Film Test"},
		},
	}

	srv := newServer()
	srv.store = testStore

	req := httptest.NewRequest(http.MethodDelete, "/api/movies/1", nil)
	r := httptest.NewRecorder()

	handler := srv.handleDeleteMovieById()
	handler.ServeHTTP(r, req)

	assert.Equal(t, http.StatusOK, r.Code, "Le statut HTTP attendu est 200 OK")

	err := testStore.DeleteMovieById(1)
	assert.Error(t, err, "Une erreur est attendue lors de la tentative de suppression d'un film non existant.")
}

func TestMovieCreateIntegration(t *testing.T) {
	srv := newServer()
	srv.store = &TestStore{}

	p := struct {
		Title       string `json:"title"`
		ReleaseDate string `json:"release_date"`
		Duration    int    `json:"duration"`
		TrailerURL  string `json:"trailer_url"`
	}{
		Title:       "Inception",
		ReleaseDate: "2010-07-18",
		Duration:    148,
		TrailerURL:  "http://url",
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(p)
	assert.Nil(t, err)

	r := httptest.NewRequest("POST", "/api/movies/", &buf)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTEzMTYyNDksImlhdCI6MTcxMTMxMjY0OSwidXNlcm5hbWUiOiJnb2xhbmcifQ.U891QTLmD5hh1QqEUVrLQ0PLAGMSHY5fhMCkIMfaZYE"
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	w := httptest.NewRecorder()

	srv.serveHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}
