package api

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

// statusRecorder запоминает код ответа, который выставил обработчик.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// Logging — логирует метод, путь, код ответа и время обработки.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		path := r.URL.Path
		if q, err := url.QueryUnescape(r.URL.RawQuery); err == nil && q != "" {
			path += "?" + q // раскодируем query, чтобы кириллица в логе читалась
		}
		log.Printf("%-6s %-20s %d %v", r.Method, path, rec.status, time.Since(start))
	})
}

// CORS — простейший CORS: разрешаем запросы с любого источника
// и сразу отвечаем 204 на preflight-запрос OPTIONS.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Router собирает все маршруты (используется и в main, и в тестах).
func Router(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)

	// Коллекция
	mux.HandleFunc("GET /tasks", h.ListTasks)
	mux.HandleFunc("POST /tasks", h.CreateTask)

	// Элемент (path-параметр {id}, Go 1.22+)
	mux.HandleFunc("GET /tasks/{id}", h.GetTask)
	mux.HandleFunc("PATCH /tasks/{id}", h.PatchTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	// Цепочка: Logging → CORS → маршрутизатор
	return Logging(CORS(mux))
}
