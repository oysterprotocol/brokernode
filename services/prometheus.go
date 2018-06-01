package services

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrepareHistogram func(name string, help string, labelNames ...string) (histogram *prometheus.HistogramVec)
type HistogramSeconds func(histogram *prometheus.HistogramVec, startTime time.Time) ()

type PrometheusService struct {
	PrepareHistogram
	HistogramSeconds
}

var (
	PrometheusWrapper PrometheusService
)

func init() {
	PrometheusWrapper = PrometheusService{
		PrepareHistogram: prepareHistogram,
		HistogramSeconds: histogramSeconds,
	}
}

func Duration(start time.Time) (duration time.Duration) {
	duration = time.Since(start)
	return duration
}

func prepareHistogram(name string, help string, labelNames ...string) (histogram *prometheus.HistogramVec) {
	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name,
		Help: help,
	}, labelNames)

	prometheus.Register(histogram)
	return histogram
}

func histogramSeconds(histogram *prometheus.HistogramVec, start time.Time) () {
	duration := Duration(start)
	histogram.WithLabelValues("500").Observe(duration.Seconds())
}
