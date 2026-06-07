package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func createStack(xs ...Middleware)Middleware{
	return func(next http.Handler) http.Handler {
		for i:=len(xs) -1 ; i>= 0 ;i--{
			next =xs[i](next)
		}
		return next
	}
}

type User struct{
	ID int `json:"id"`
	Name string `json:"name"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int){
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw:=&responseWriter{ResponseWriter: w,statusCode: http.StatusOK}
		log.Println("->",r.Method, r.URL.Path)
		next.ServeHTTP(rw,r)
		log.Println("<-",r.Method,r.URL.Path,time.Since(start))
	})
}

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "hello")
		next.ServeHTTP(w, r)
	})
}

func main(){
	mux:= http.NewServeMux()
	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(tokenKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"id":    r.PathValue("id"),
		"token": token,
	})
})

	mux.HandleFunc("GET /set-cookie", func(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: "abc123",
		Path:  "/",
	})
	w.Write([]byte("cookie set"))
})

mux.HandleFunc("GET /get-cookie", func(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "no cookie", http.StatusNotFound)
		return
	}
	w.Write([]byte("cookie: " + cookie.Value))
})
	mux.Handle("/users/",http.StripPrefix("/users",userRouter()))
	mux.Handle("/todos/",http.StripPrefix("/todos",todoRouter()))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	mux.HandleFunc("GET api.example.com/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello from api subdomain"))
})

	chain:= createStack(loggingMiddleware,headerMiddleware,authMiddleware)
	handler := chain(mux)
	log.Println("Listening on :443")
	http.ListenAndServeTLS(":443", "server.crt","server.key",handler)
}

/*
generated certs from this 
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key -out server.crt \
  -subj "/CN=localhost"

*/