package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/alioygur/gores"
	"github.com/go-chi/chi/v5"
)

type Middleware func(http.Handler) http.Handler

var ErrNotFound = errors.New("not found")

func MakeURLParamMiddleware[T any](param string, getter func(value string) (*T, error)) (Middleware, func(r *http.Request) *T) {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				value := chi.URLParam(r, param)
				if value == "" {
					gores.Error(w, http.StatusNotFound, "not found")
					return
				}

				inst, err := getter(value)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						gores.Error(w, http.StatusNotFound, "not found")
						return
					}
					gores.Error(w, http.StatusInternalServerError, "failed to process url paramater")
					return
				}

				ctx := context.WithValue(r.Context(), param, inst)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}, func(r *http.Request) *T {
			return r.Context().Value(param).(*T)
		}
}
