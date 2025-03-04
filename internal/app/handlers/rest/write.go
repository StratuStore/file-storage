package rest

import (
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

// WriteFile is a POST request
// Body must be a file
func (h *Handler) WriteFile(w http.ResponseWriter, req *http.Request) {
	l := h.l.With(slog.String("op", "internal.app.handlers.rest.WriteFile"))

	rawConnectionID := req.URL.Query().Get("connectionID")

	connectionID, err := uuid.Parse(rawConnectionID)
	if err != nil {
		l.Debug("unable to decode request query", slog.String("err", err.Error()))
		_ = h.handleError(w, http.StatusBadRequest, err, "invalid connectionID")
		return
	}

	if req.Body == nil {
		l.Debug("request body is empty")
		_ = h.handleError(w, http.StatusBadRequest, err, "request body is empty")
		return
	}

	err = h.useCases.Write(req.Context(), connectionID, req.Body)
	if err != nil {
		_ = h.handleError(w, http.StatusNotFound, err, "connection error")
		l.Debug("unable to find suitable file reader", slog.String("err", err.Error()))
		return
	}
}
