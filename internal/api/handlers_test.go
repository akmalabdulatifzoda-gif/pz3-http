package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/pz3-http/internal/storage"
)

func newTestRouter() http.Handler {
	return Router(NewHandlers(storage.NewMemoryStore()))
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(t, newTestRouter(), "GET", "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q", ct)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestCreateTask(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"ok", `{"title":"Купить молоко"}`, http.StatusCreated},
		{"empty title", `{"title":"   "}`, http.StatusBadRequest},
		{"invalid json", `{"title":`, http.StatusBadRequest},
		{"too short", `{"title":"ab"}`, http.StatusUnprocessableEntity},
		{"too long", `{"title":"` + strings.Repeat("я", 141) + `"}`, http.StatusUnprocessableEntity},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := do(t, newTestRouter(), "POST", "/tasks", c.body)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, c.want, rec.Body.String())
			}
		})
	}
}

func TestTaskLifecycle(t *testing.T) {
	h := newTestRouter()

	rec := do(t, h, "POST", "/tasks", `{"title":"Сделать ПЗ №3"}`)
	var created storage.Task
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID != 1 || rec.Header().Get("Location") != "/tasks/1" {
		t.Fatalf("unexpected create result: %+v, Location=%q", created, rec.Header().Get("Location"))
	}

	if rec := do(t, h, "GET", "/tasks/1", ""); rec.Code != http.StatusOK {
		t.Fatalf("GET existing: %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/tasks/abc", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("GET abc: %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/tasks/9999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET 9999: %d", rec.Code)
	}

	rec = do(t, h, "PATCH", "/tasks/1", `{"done":true}`)
	var patched storage.Task
	_ = json.Unmarshal(rec.Body.Bytes(), &patched)
	if rec.Code != http.StatusOK || !patched.Done {
		t.Fatalf("PATCH: %d %+v", rec.Code, patched)
	}

	if rec := do(t, h, "DELETE", "/tasks/1", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/tasks/1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET after delete: %d", rec.Code)
	}
}

func TestListFilter(t *testing.T) {
	h := newTestRouter()
	do(t, h, "POST", "/tasks", `{"title":"Купить молоко"}`)
	do(t, h, "POST", "/tasks", `{"title":"Сделать ПЗ"}`)

	rec := do(t, h, "GET", "/tasks?q=МОЛОКО", "")
	var tasks []storage.Task
	_ = json.Unmarshal(rec.Body.Bytes(), &tasks)
	if len(tasks) != 1 || tasks[0].Title != "Купить молоко" {
		t.Fatalf("filter result: %+v", tasks)
	}
}

func TestCORS(t *testing.T) {
	rec := do(t, newTestRouter(), "OPTIONS", "/tasks", "")
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("preflight: %d %v", rec.Code, rec.Header())
	}
}
