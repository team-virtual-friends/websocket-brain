package foundation

import (
	"time"
)

var (
	globalMetricsClient Metrics = NewMetricsClient()
)

type Metrics interface {
	RecordCount(name string, count int64)
	RecordLevel(name string, level int64)
	RecordGauge(name string, gauge float64)
	RecordDuration(name string, duration time.Duration)
	RecordDurationSince(name string, start time.Time)

	Close()
}

type MetricsClient struct {
}

func NewMetricsClient() Metrics {
	return &MetricsClient{}
}

func (t *MetricsClient) RecordCount(name string, count int64) {

}

func (t *MetricsClient) RecordLevel(name string, level int64) {

}

func (t *MetricsClient) RecordGauge(name string, gauge float64) {

}

func (t *MetricsClient) RecordDuration(name string, duration time.Duration) {

}

func (t *MetricsClient) RecordDurationSince(name string, start time.Time) {

}

func (t *MetricsClient) Close() {

}
