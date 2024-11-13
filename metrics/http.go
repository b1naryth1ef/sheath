package metrics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		httpRequests.WithLabelValues(r.Method, routePattern).Inc()
	})
}

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_http_requests",
		Help: "The total number of http requests",
	}, []string{"method", "route"})
)
