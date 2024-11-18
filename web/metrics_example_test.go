package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/b1naryth1ef/sheath/web"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
)

func ExampleMetricsMiddleware() {
	rtr := chi.NewRouter()
	rtr.Use(web.MetricsMiddleware)
	rtr.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	rtr.Handle("/metrics", promhttp.Handler())

	// Execute a request so we increment our metrics
	r := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, r)

	// Fetch the latest prometheus metrics
	r = httptest.NewRequest("GET", "http://example.com/metrics", nil)
	w = httptest.NewRecorder()
	rtr.ServeHTTP(w, r)

	// Parse the response in expfmt
	resp := w.Result()
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		panic(err)
	}

	entry, ok := mf["go_http_request_count"]
	if !ok {
		panic("go_http_request_count metric not found in prom response")
	}
	fmt.Printf("%s %0.0f\n", *entry.Metric[0].Label[2].Value, *entry.Metric[0].Counter.Value)
	// Output: 204 1
}
