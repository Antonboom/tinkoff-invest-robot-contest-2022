package main

import (
	"errors"
	stdlog "log"
	"net/http"
)
import "github.com/prometheus/client_golang/prometheus/promhttp"

func runMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			stdlog.Println("cannot run metrics exposure server")
		}
	}()
}
