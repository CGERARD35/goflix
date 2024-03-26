package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Goflix")
	}
}

func (s *server) handleTokenCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type response struct {
		Token string `json:"token"`
	}

	type responseError struct {
		Error string `json:"error"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		err := s.decode(w, r, &req)
		if err != err {
			msg := fmt.Sprintf("Cannot parse login body. err=%v\n", err)
			log.Println(msg)
			s.respond(w, r, responseError{
				Error: msg,
			}, http.StatusBadRequest)
			return
		}

		found, err := s.store.FindUser(req.Username, req.Password)
		if err != nil {
			msg := fmt.Sprintf("Cannot find user. err=%v", err)
			s.respond(w, r, responseError{
				Error: msg,
			}, http.StatusInternalServerError)
			return
		}
		if !found {
			s.respond(w, r, responseError{
				Error: "Invalid credentials",
			}, http.StatusInternalServerError)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": req.Username,
			"exp":      time.Now().Add(time.Hour * time.Duration(1)).Unix(),
			"iat":      time.Now().Unix(),
		})

		tokenStr, err := token.SignedString([]byte(JWT_APP_KEY))
		if err != nil {
			msg := fmt.Sprintf("Cannot generate JWT. err=%v", err)
			s.respond(w, r, responseError{
				Error: msg,
			}, http.StatusInternalServerError)
			return
		}
		s.respond(w, r, response{
			Token: tokenStr,
		}, http.StatusOK)
	}
}
