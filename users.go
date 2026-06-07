package main

import (
	"encoding/json"
	"net/http"
)

func userRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("list users"))
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create user"))
	})

	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": r.PathValue("id")})
	})

	return mux
}