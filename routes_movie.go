package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type jsonMovie struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Duration    int    `json:"duration"`
	TrailerURL  string `json:"trailer_url"`
}

func (s *server) handleMovieList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movies, err := s.store.GetMovies()
		if err != nil {
			log.Printf("Cannot load movies. err=%v\n", err)
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("movies.html")
		if err != nil {
			log.Printf("Error loading template: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, movies)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func (s *server) handleMovieDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			log.Printf("Cannot parse id to int. err=%v", err)
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}
		m, err := s.store.GetMovieById(id)
		if err != nil {
			log.Printf("Cannot load movie. error=%v\n", err)
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		var resp = mapMovieToJson(m)
		s.respond(w, r, resp, http.StatusOK)
	}
}

func (s *server) handleFindMovieById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}
		movie, err := s.store.GetMovieById(id) // Assurez-vous d'utiliser GetMovieById ici
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Movie not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to find the movie", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(movie)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (s *server) handleMovieCreate() http.HandlerFunc {
	type request struct {
		Title       string `json:"title"`
		ReleaseDate string `json:"release_date"`
		Duration    int    `json:"duration"`
		TrailerURL  string `json:"trailer_url"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		err := s.decode(w, r, &req)
		if err != nil {
			log.Printf("Cannot parse movie body. erreur=%v\n", err)
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}
		m := &Movie{
			ID:          0,
			Title:       req.Title,
			ReleaseDate: req.ReleaseDate,
			Duration:    req.Duration,
			TrailerURL:  req.TrailerURL,
		}

		err = s.store.CreateMovie(m)
		if err != nil {
			log.Printf("Cannot create movie in DB. err=%v", err)
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		var resp = mapMovieToJson(m)
		s.respond(w, r, resp, http.StatusOK)
	}
}

func (s *server) handleUpdateMovie() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		var movie Movie
		if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.store.UpdateMovie(id, movie)
		if err != nil {
			http.Error(w, "Failed to update movie", http.StatusInternalServerError)
			return
		}

		updatedMovie, err := s.store.GetMovieById(id)
		if err != nil {
			http.Error(w, "Failed to retrieve updated movie", http.StatusInternalServerError)
			return
		}

		s.respond(w, r, updatedMovie, http.StatusOK)
	}
}

func (s *server) handleDeleteMovieById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}
		err = s.store.DeleteMovieById(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Movie not found", http.StatusNotFound)
			} else {
				log.Printf("Failed to delete movie: %v", err)
				http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func mapMovieToJson(m *Movie) jsonMovie {
	return jsonMovie{
		ID:          m.ID,
		Title:       m.Title,
		ReleaseDate: m.ReleaseDate,
		Duration:    m.Duration,
		TrailerURL:  m.TrailerURL,
	}
}
