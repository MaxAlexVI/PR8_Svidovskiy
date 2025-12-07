package notes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ repo *Repo }

func NewHandler(r *Repo) *Handler { return &Handler{repo: r} }

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.del)
	r.Get("/stats", h.stats) // Новый эндпоинт для статистики
	return r
}

func reqCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title     string     `json:"title"`
		Content   string     `json:"content"`
		ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
		writeJSON(w, 400, map[string]string{"error":"invalid_json_or_title"})
		return
	}
	c, cancel := reqCtx(r); defer cancel()
	n, err := h.repo.Create(c, in.Title, in.Content, in.ExpiresAt)
	if err != nil { writeJSON(w, 409, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 201, n)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, cancel := reqCtx(r); defer cancel()
	n, err := h.repo.ByID(c, id)
	if errors.Is(err, ErrNotFound) { writeJSON(w, 404, map[string]string{"error":"not_found"}); return }
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, n)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	skip,  _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)
	after := r.URL.Query().Get("after") // Параметр для курсорной пагинации
	
	// Используем курсорную пагинацию если передан after, иначе обычную
	if after != "" {
		if limit <= 0 || limit > 200 { limit = 20 }
		
		c, cancel := reqCtx(r); defer cancel()
		items, err := h.repo.ListCursor(c, q, limit, after)
		if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
		
		// Добавляем информацию о следующей странице
		response := map[string]interface{}{
			"notes": items,
			"hasMore": len(items) == int(limit),
		}
		if len(items) > 0 {
			response["nextCursor"] = items[len(items)-1].ID.Hex()
		}
		writeJSON(w, 200, response)
		return
	}
	
	// Обычная пагинация
	if limit <= 0 || limit > 200 { limit = 20 }
	if skip < 0 { skip = 0 }

	c, cancel := reqCtx(r); defer cancel()
	items, err := h.repo.List(c, q, limit, skip)
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, items)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in struct {
		Title     *string    `json:"title"`
		Content   *string    `json:"content"`
		ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error":"invalid_json"})
		return
	}
	c, cancel := reqCtx(r); defer cancel()
	n, err := h.repo.Update(c, id, in.Title, in.Content, in.ExpiresAt)
	if errors.Is(err, ErrNotFound) { writeJSON(w, 404, map[string]string{"error":"not_found"}); return }
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, n)
}

func (h *Handler) del(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, cancel := reqCtx(r); defer cancel()
	if err := h.repo.Delete(c, id); errors.Is(err, ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error":"not_found"}); return
	} else if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()}); return
	}
	w.WriteHeader(204)
}

// stats - новый обработчик для получения статистики
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	c, cancel := reqCtx(r); defer cancel()
	
	stats, err := h.repo.GetStats(c)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	
	writeJSON(w, 200, stats)
}