package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"autotrade/backend-go/internal/model"
	"autotrade/backend-go/internal/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() http.Handler {
	r := mux.NewRouter()
	r.Use(cors)

	r.HandleFunc("/healthz", h.health).Methods(http.MethodGet)
	r.HandleFunc("/v1/me/state", h.getState).Methods(http.MethodGet)
	r.HandleFunc("/v1/me/fills", h.getFills).Methods(http.MethodGet)
	r.HandleFunc("/v1/me/review", h.postReview).Methods(http.MethodPost)
	r.HandleFunc("/v1/strategy/derives", h.getDerives).Methods(http.MethodGet)
	r.HandleFunc("/v1/control/bias", h.patchBias).Methods(http.MethodPatch)
	return r
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getState(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.State(r.Context())
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

func (h *Handler) getFills(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	fills, err := h.svc.Fills(r.Context(), limit)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"fills": fills})
}

func (h *Handler) postReview(w http.ResponseWriter, r *http.Request) {
	var in model.ReviewInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondErr(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.Review(r.Context(), in); err != nil {
		if service.IsBadRequest(err) {
			respondErr(w, http.StatusBadRequest, err)
			return
		}
		respondErr(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"status": "saved"})
}

func (h *Handler) getDerives(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Derives(r.Context())
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"derives": items})
}

func (h *Handler) patchBias(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Bias string `json:"bias"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondErr(w, http.StatusBadRequest, err)
		return
	}
	if err := h.svc.SetBias(r.Context(), body.Bias); err != nil {
		if service.IsBadRequest(err) {
			respondErr(w, http.StatusBadRequest, err)
			return
		}
		respondErr(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func respondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondErr(w http.ResponseWriter, code int, err error) {
	respondJSON(w, code, map[string]string{"error": err.Error()})
}
