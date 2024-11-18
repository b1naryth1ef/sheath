package web

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alioygur/gores"
	"gorm.io/gorm"
)

// DBErrorResponse returns either a 404 or 500 response based on the given error.
func DBErrorResponse(w http.ResponseWriter, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		gores.Error(w, http.StatusNotFound, "not found")
		return
	}

	slog.Error("db error response", slog.Any("err", err))
	gores.Error(w, http.StatusInternalServerError, "something went wrong")
}
