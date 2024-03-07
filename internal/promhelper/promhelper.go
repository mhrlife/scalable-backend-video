package promhelper

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	StatusOk       Status = "ok"
	StatusError    Status = "error"
	StatusNotFound Status = "not-found"
)

type HistogramWithCounter struct {
	histogram *prometheus.HistogramVec
	counter   *prometheus.CounterVec
}

func NewHistogramWithCounter(name string, buckets []float64) *HistogramWithCounter {
	histogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name + "_histogram",
		Help:    "auto generated " + name + " histogram",
		Buckets: buckets,
	}, []string{"task_name", "status"})
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name + "_counter",
		Help: "auto generated " + name + " counter",
	}, []string{"task_name", "status"})

	return &HistogramWithCounter{
		histogram: histogram,
		counter:   counter,
	}
}

func (h *HistogramWithCounter) Do(taskName string, task func() error) error {
	s := time.Now()
	err := task()
	dur := time.Since(s)
	status := StatusOk
	if err != nil {
		status = StatusError
	}
	var promError PromError
	if errors.As(err, &promError) {
		status = promError.status
		err = promError.error
	}

	h.histogram.WithLabelValues(taskName, status.String()).Observe(dur.Seconds())
	h.counter.WithLabelValues(taskName, status.String()).Inc()

	return err
}
