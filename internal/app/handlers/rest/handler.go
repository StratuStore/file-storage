package rest

import (
	"encoding/json"
	"github.com/StratuStore/file-storage/internal/app/usecases"
	"log/slog"
	"net/http"
)

type Handler struct {
	useCases usecases.UseCases
	l        *slog.Logger
}

func NewHandler(useCases usecases.UseCases, logger *slog.Logger) *Handler {
	return &Handler{
		useCases: useCases,
		l:        logger.With(slog.String("op", "internal.app.handlers.rest.Handler")),
	}
}

func (h *Handler) handleError(w http.ResponseWriter, status int, message string) error {
	w.WriteHeader(status)
	response := ResponseError{Err: message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.l.Error("unable to encode response", slog.String("err", err.Error()))
		return err
	}

	return nil
}

type ResponseError struct {
	Err string `json:"error"`
}

func (e *ResponseError) Error() string {
	return e.Err
}
