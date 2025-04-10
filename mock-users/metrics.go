package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mock_request_duration_seconds",
			Help:    "Время отклика API",
			Buckets: prometheus.LinearBuckets(0.05, 0.1, 20),
		},
		[]string{"endpoint"},
	)

	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_requests_total",
			Help: "Общее число запросов к API",
		},
		[]string{"endpoint"},
	)
)

var (
	requestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_request_errors_total",
			Help: "Число ошибок (status >= 400)",
		},
		[]string{"endpoint"},
	)
)
var waveNumber = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "mock_wave_number",
	Help: "Текущий номер волны нагрузки",
})

func registerMetrics() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(requestErrors)
	prometheus.MustRegister(waveNumber)
}
