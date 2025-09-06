package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var RequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bytesize_requests_total",
		Help: "Total service requests by endpoint.",
	},
	[]string{"endpoint"},
)

var ErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bytesize_errors_total",
		Help: "Total errors by endpoint.",
	},
	[]string{"endpoint"},
)

var RequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "bytesize_request_duration_seconds",
		Help:    "Service duration by endpoint.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"endpoint"},
)

var BytesUploadedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "bytesize_bytes_uploaded_total",
		Help: "Sum of uploaded bytes on successful uploads.",
	},
)

var BytesStreamedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "bytesize_bytes_streamed_total",
		Help: "Sum of streamed bytes on successful downloads.",
	},
)
