package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"example.com/pz3-http/internal/storage"
)

const (
	minTitleLen = 3
	maxTitleLen = 140
	maxBodySize = 1 << 20 // 1 МБ — защита от слишком больших тел запроса
)

type Handlers struct {
	Store *storage.MemoryStore
}

func NewHandlers(store *storage.MemoryStore) *Handlers {
	return &Handlers{Store: store}
}

// Health — GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListTasks — GET /tasks, фильтр ?q=text (без учёта регистра)
func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.Store.List()

	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if q != "" {
		filtered := tasks[:0]
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), q) {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}
	JSON(w, http.StatusOK, tasks)
}

type createTaskRequest struct {
	Title string `json:"title"`
}

// CreateTask — POST /tasks, тело {"title":"..."}
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		BadRequest(w, "Content-Type must be application/json")
		return
	}
	var req createTaskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		BadRequest(w, "title is required")
		return
	}
	if n := utf8.RuneCountInString(title); n < minTitleLen || n > maxTitleLen {
		Unprocessable(w, "title length must be between 3 and 140 characters")
		return
	}
	t := h.Store.Create(title)
	w.Header().Set("Location", "/tasks/"+strconv.FormatInt(t.ID, 10))
	JSON(w, http.StatusCreated, t)
}

// GetTask — GET /tasks/{id}
func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Get(id)
	if err != nil {
		storeError(w, err)
		return
	}
	JSON(w, http.StatusOK, t)
}

type patchTaskRequest struct {
	Done *bool `json:"done"` // указатель — чтобы отличить «не передано» от false
}

// PatchTask — PATCH /tasks/{id}, тело {"done":true}
func (h *Handlers) PatchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if !isJSON(r) {
		BadRequest(w, "Content-Type must be application/json")
		return
	}
	var req patchTaskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Done == nil {
		BadRequest(w, "field done is required")
		return
	}
	t, err := h.Store.SetDone(id, *req.Done)
	if err != nil {
		storeError(w, err)
		return
	}
	JSON(w, http.StatusOK, t)
}

// DeleteTask — DELETE /tasks/{id}, успешный ответ 204 без тела
func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(id); err != nil {
		storeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- вспомогательные функции ----------

// parseID читает path-параметр {id} (Go 1.22+: r.PathValue) и проверяет, что это положительное число.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		BadRequest(w, "invalid id")
		return 0, false
	}
	return id, true
}

func isJSON(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return ct == "" || strings.Contains(ct, "application/json")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		BadRequest(w, "invalid json: "+err.Error())
		return false
	}
	return true
}

func storeError(w http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		NotFound(w, "task not found")
		return
	}
	Internal(w, "unexpected error")
}
