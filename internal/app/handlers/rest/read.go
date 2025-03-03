package rest

import (
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

// ReadFile is a GET request
func (h *Handler) ReadFile(w http.ResponseWriter, req *http.Request) {
	l := h.l.With(slog.String("op", "internal.app.handlers.rest.ReadFile"))

	rawConnectionID := req.URL.Query().Get("connectionID")

	connectionID, err := uuid.Parse(rawConnectionID)
	if err != nil {
		l.Debug("unable to decode request query", slog.String("err", err.Error()))
		_ = h.handleError(w, http.StatusBadRequest, "invalid connectionID")
		return
	}

	reader, err := h.useCases.Read(req.Context(), connectionID)
	if err != nil {
		_ = h.handleError(w, http.StatusNotFound, "connection not found")
		l.Debug("unable to find suitable file reader", slog.String("err", err.Error()))
		return
	}

	http.ServeContent(w, req, "", time.Time{}, reader)
}
