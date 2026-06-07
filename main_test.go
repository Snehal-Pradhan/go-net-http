package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- createStack ---

func TestCreateStack_callsAllInOrder(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		})
	}

	stack := createStack(mw1, mw2)
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	stack(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d: %v", len(order), order)
	}
	if order[0] != "mw1" || order[1] != "mw2" || order[2] != "handler" {
		t.Errorf("wrong order: %v", order)
	}
}

func TestCreateStack_empty(t *testing.T) {
	stack := createStack()
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	w := httptest.NewRecorder()
	stack(final).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %q", w.Body.String())
	}
}

// --- responseWriter ---

func TestResponseWriter_capturesStatusCode(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder(), statusCode: http.StatusOK}
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rw.statusCode)
	}
}

// --- headerMiddleware ---

func TestHeaderMiddleware_setsCustomHeader(t *testing.T) {
	mw := headerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	if h := w.Header().Get("X-Custom"); h != "hello" {
		t.Errorf("expected X-Custom: hello, got %q", h)
	}
}

// --- authMiddleware ---

func TestAuthMiddleware_extractsBearerToken(t *testing.T) {
	var gotToken string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(tokenKey).(string)
		if ok {
			gotToken = token
		}
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	authMiddleware(next).ServeHTTP(httptest.NewRecorder(), req)

	if gotToken != "test-token-123" {
		t.Errorf("expected 'test-token-123', got %q", gotToken)
	}
}

func TestAuthMiddleware_noHeader_doesNotSetToken(t *testing.T) {
	var tokenSet bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, tokenSet = r.Context().Value(tokenKey).(string)
	})
	authMiddleware(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if tokenSet {
		t.Error("token should not be set without auth header")
	}
}

// --- userRouter ---

func TestUserRouter_listUsers(t *testing.T) {
	w := httptest.NewRecorder()
	userRouter().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Body.String() != "list users" {
		t.Errorf("expected 'list users', got %q", w.Body.String())
	}
}

func TestUserRouter_createUser(t *testing.T) {
	w := httptest.NewRecorder()
	userRouter().ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Body.String() != "create user" {
		t.Errorf("expected 'create user', got %q", w.Body.String())
	}
}

func TestUserRouter_getByID(t *testing.T) {
	w := httptest.NewRecorder()
	userRouter().ServeHTTP(w, httptest.NewRequest("GET", "/42", nil))
	if w.Body.String() != `{"id":"42"}`+"\n" {
		t.Errorf("expected '{\"id\":\"42\"}\\n', got %q", w.Body.String())
	}
}

// --- full handler test (via chain) ---

func TestHandler_home(t *testing.T) {
	handler := createStack(loggingMiddleware, headerMiddleware, authMiddleware)(muxForTest())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Body.String() != "Home" {
		t.Errorf("expected 'Home', got %q", w.Body.String())
	}
}

func TestHandler_setAndGetCookie(t *testing.T) {
	handler := createStack(loggingMiddleware, headerMiddleware, authMiddleware)(muxForTest())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/set-cookie", nil)
	handler.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" {
		t.Fatalf("expected session cookie, got %v", cookies)
	}
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/get-cookie", nil)
	req2.AddCookie(cookies[0])
	handler.ServeHTTP(w2, req2)

	if w2.Body.String() != "cookie: abc123" {
		t.Errorf("expected 'cookie: abc123', got %q", w2.Body.String())
	}
}

func TestHandler_withBearerToken(t *testing.T) {
	handler := createStack(loggingMiddleware, headerMiddleware, authMiddleware)(muxForTest())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/42", nil)
	req.Header.Set("Authorization", "Bearer my-token")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandler_noAuth_returnsUnauthorized(t *testing.T) {
	handler := createStack(loggingMiddleware, headerMiddleware, authMiddleware)(muxForTest())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/42", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// muxForTest recreates the mux without ListenAndServeTLS
func muxForTest() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(tokenKey).(string)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Write([]byte("id=" + r.PathValue("id") + " token=" + token))
	})
	mux.HandleFunc("GET /set-cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/"})
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
	mux.Handle("/users/", http.StripPrefix("/users", userRouter()))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})
	mux.HandleFunc("GET api.example.com/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from api subdomain"))
	})
	return mux
}
