package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsMiddleware is an HTTP middleware that tracks chi/v5 requests with Prometheus
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		httpRequests.WithLabelValues(r.Method, routePattern, strconv.Itoa(ww.Status())).Inc()
		httpRequestLatency.WithLabelValues(r.Method, routePattern, strconv.Itoa(ww.Status())).Observe(time.Since(start).Seconds())
	})
}

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_http_request_count",
		Help: "The total number of http requests",
	}, []string{"method", "route", "status"})

	httpRequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "go_http_request_duration_seconds",
		Help:    "Histogram of request latency in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "route", "status"})
)
