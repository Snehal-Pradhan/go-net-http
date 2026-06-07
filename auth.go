package main

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const tokenKey contextKey = "token"

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			ctx := context.WithValue(r.Context(), tokenKey, token)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}