package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func runMetrics(ctx context.Context, addr string) error {
	http.Handle("/metrics", promhttp.Handler())

	s := &http.Server{Addr: addr}

	go func() {
		<-ctx.Done()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel2()
		_ = s.Shutdown(ctx2)
	}()

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("run metrics exposure server: %v", err)
	}
	return nil
}
