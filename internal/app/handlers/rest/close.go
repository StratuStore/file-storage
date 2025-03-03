package rest

import (
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

// CloseFile is a POST request
func (h *Handler) CloseFile(w http.ResponseWriter, req *http.Request) {
	l := h.l.With(slog.String("op", "internal.app.handlers.rest.CloseFile"))

	rawConnectionID := req.URL.Query().Get("connectionID")

	connectionID, err := uuid.Parse(rawConnectionID)
	if err != nil {
		l.Debug("unable to decode request query", slog.String("err", err.Error()))
		_ = h.handleError(w, http.StatusBadRequest, "invalid connectionID")
		return
	}

	err = h.useCases.Close(req.Context(), connectionID)
	if err != nil {
		_ = h.handleError(w, http.StatusNotFound, "connection error")
		l.Debug("unable to find suitable file reader", slog.String("err", err.Error()))
		return
	}
}
