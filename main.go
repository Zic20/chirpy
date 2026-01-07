package main

import (
	"net/http"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler{})
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	s.ListenAndServe()
}
