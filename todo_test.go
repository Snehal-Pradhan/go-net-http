package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resetTodoStore() {
	todoStore = NewTodoStore()
}

func withToken(req *http.Request) *http.Request {
	req.Header.Set("Authorization", "Bearer test-token")
	return req
}

func todoWithAuth() http.Handler {
	return authMiddleware(todoRouter())
}

func TestTodoList_empty(t *testing.T) {
	resetTodoStore()
	w := httptest.NewRecorder()
	todoRouter().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "[]\n" {
		t.Errorf("expected '[]\\n', got %q", w.Body.String())
	}
}

func TestTodoCreate_withAuth(t *testing.T) {
	resetTodoStore()
	body := `{"title":"buy milk"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var todo Todo
	if err := json.NewDecoder(w.Body).Decode(&todo); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if todo.ID != 1 {
		t.Errorf("expected ID 1, got %d", todo.ID)
	}
	if todo.Title != "buy milk" {
		t.Errorf("expected 'buy milk', got %q", todo.Title)
	}
	if todo.Completed {
		t.Errorf("expected completed=false")
	}
	if todo.CreatedAt.IsZero() {
		t.Errorf("expected non-zero CreatedAt")
	}
}

func TestTodoCreate_noAuth(t *testing.T) {
	resetTodoStore()
	body := `{"title":"buy milk"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTodoCreate_emptyTitle(t *testing.T) {
	resetTodoStore()
	body := `{"title":""}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTodoCreate_invalidJSON(t *testing.T) {
	resetTodoStore()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTodoGetByID(t *testing.T) {
	resetTodoStore()
	todoStore.Create("test todo")

	w := httptest.NewRecorder()
	todoRouter().ServeHTTP(w, httptest.NewRequest("GET", "/1", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var todo Todo
	json.NewDecoder(w.Body).Decode(&todo)
	if todo.Title != "test todo" {
		t.Errorf("expected 'test todo', got %q", todo.Title)
	}
}

func TestTodoGetByID_notFound(t *testing.T) {
	resetTodoStore()
	w := httptest.NewRecorder()
	todoRouter().ServeHTTP(w, httptest.NewRequest("GET", "/999", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTodoGetByID_invalidID(t *testing.T) {
	resetTodoStore()
	w := httptest.NewRecorder()
	todoRouter().ServeHTTP(w, httptest.NewRequest("GET", "/abc", nil))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTodoUpdate_withAuth(t *testing.T) {
	resetTodoStore()
	todoStore.Create("old title")

	body := `{"title":"new title","completed":true}`
	req := httptest.NewRequest("PUT", "/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var todo Todo
	json.NewDecoder(w.Body).Decode(&todo)
	if todo.Title != "new title" {
		t.Errorf("expected 'new title', got %q", todo.Title)
	}
	if !todo.Completed {
		t.Errorf("expected completed=true")
	}
}

func TestTodoUpdate_noAuth(t *testing.T) {
	resetTodoStore()
	todoStore.Create("old title")

	body := `{"title":"new title"}`
	req := httptest.NewRequest("PUT", "/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTodoUpdate_notFound(t *testing.T) {
	resetTodoStore()
	body := `{"title":"new","completed":false}`
	req := httptest.NewRequest("PUT", "/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTodoDelete_withAuth(t *testing.T) {
	resetTodoStore()
	todoStore.Create("to delete")

	req := httptest.NewRequest("DELETE", "/1", nil)
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestTodoDelete_noAuth(t *testing.T) {
	resetTodoStore()
	todoStore.Create("to delete")

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, httptest.NewRequest("DELETE", "/1", nil))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTodoDelete_notFound(t *testing.T) {
	resetTodoStore()
	req := httptest.NewRequest("DELETE", "/999", nil)
	req = withToken(req)

	w := httptest.NewRecorder()
	todoWithAuth().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTodoList_afterCreate(t *testing.T) {
	resetTodoStore()
	todoStore.Create("first")
	todoStore.Create("second")

	w := httptest.NewRecorder()
	todoRouter().ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	var todos []Todo
	json.NewDecoder(w.Body).Decode(&todos)
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	if todos[0].Title != "first" {
		t.Errorf("expected 'first', got %q", todos[0].Title)
	}
	if todos[1].Title != "second" {
		t.Errorf("expected 'second', got %q", todos[1].Title)
	}
}
