package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("GoFlix")

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	srv := newServer()
	srv.store = &dbStore{}
	err := srv.store.Open()
	if err != nil {
		return err
	}
	defer srv.store.Close()

	http.HandleFunc("/", srv.serveHTTP)
	log.Printf("Serving HTTP on port 9000")
	err = http.ListenAndServe(":9000", nil)
	if err != nil {
		return err
	}
	return nil
}
