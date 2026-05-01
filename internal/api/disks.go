package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/example/diskhm/internal/domain"
)

type sleepAfterRequest struct {
	Minutes int `json:"minutes"`
}

func registerDiskRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /api/disks", requireSession(http.HandlerFunc(listDisksHandler(deps))))
	mux.Handle("POST /api/disks/{id}/sleep-now", requireSession(http.HandlerFunc(sleepNowHandler(deps))))
	mux.Handle("POST /api/disks/{id}/sleep-after", requireSession(http.HandlerFunc(sleepAfterHandler(deps))))
	mux.Handle("POST /api/disks/{id}/cancel-sleep", requireSession(http.HandlerFunc(cancelSleepHandler(deps))))
	mux.Handle("POST /api/disks/{id}/refresh-safe", requireSession(http.HandlerFunc(refreshSafeHandler(deps))))
	mux.Handle("POST /api/disks/{id}/refresh-wake", requireSession(http.HandlerFunc(refreshWakeHandler(deps))))
}

func listDisksHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disks, err := deps.Runtime.ListDisks(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to list disks")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": disks})
	}
}

func sleepNowHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.Runtime.SleepNow(r.Context(), r.PathValue("id")); err != nil {
			writeDiskActionError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func sleepAfterHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req sleepAfterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_json", "invalid json")
			return
		}
		if req.Minutes <= 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_minutes", "minutes must be greater than zero")
			return
		}

		if err := deps.Runtime.SleepAfter(r.Context(), r.PathValue("id"), req.Minutes); err != nil {
			writeDiskActionError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func cancelSleepHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.Runtime.CancelSleep(r.Context(), r.PathValue("id")); err != nil {
			writeDiskActionError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func refreshSafeHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.Runtime.RefreshSafe(r.Context(), r.PathValue("id")); err != nil {
			writeDiskActionError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func refreshWakeHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deps.Runtime.RefreshWake(r.Context(), r.PathValue("id")); err != nil {
			writeDiskActionError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{
		Code:    code,
		Message: message,
	})
}

func writeDiskActionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrDiskNotFound):
		writeAPIError(w, http.StatusNotFound, "disk_not_found", err.Error())
	case errors.Is(err, domain.ErrTaskConflict):
		writeAPIError(w, http.StatusConflict, "task_conflict", err.Error())
	case errors.Is(err, domain.ErrUnsupportedDevice):
		writeAPIError(w, http.StatusBadRequest, "unsupported_device", err.Error())
	default:
		writeAPIError(w, http.StatusInternalServerError, "command_failed", err.Error())
	}
}
