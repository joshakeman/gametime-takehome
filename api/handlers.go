package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/joshakeman/gametime-take-home/data"
)

// OrderService is what the handlers need. It's defined here (not in service)
// so api doesn't import service, which avoids an import cycle.
type OrderService interface {
	Create(userID string, amount int) (data.Order, error)
	Process(id int) (data.Order, error)
	Get(id int) (data.Order, error)
}

type Handler struct {
	svc OrderService
}

func NewHandler(svc OrderService) http.Handler {
	h := Handler{svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.create)
	mux.HandleFunc("POST /orders/{id}/process", h.process)
	mux.HandleFunc("GET /orders/{id}", h.get)
	return mux
}

func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var req Order
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.UserID == "" || req.Amount <= 0 {
		http.Error(w, "user_id and a positive amount are required", http.StatusBadRequest)
		return
	}
	o, err := h.svc.Create(req.UserID, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

// process runs authorize -> complete (-> void). A decline or failed
// completion is a recorded outcome, not a failed request, so it returns 200
// with the order's final state and the error text.
func (h Handler) process(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	o, err := h.svc.Process(id)
	if errors.Is(err, data.ErrNotFound) || errors.Is(err, data.ErrInvalidTransition) {
		writeError(w, err)
		return
	}
	resp := struct {
		Order data.Order `json:"order"`
		Error string     `json:"error,omitempty"`
	}{Order: o}
	if err != nil {
		resp.Error = err.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	o, err := h.svc.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, data.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, data.ErrInvalidTransition):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
