package histogram

import (
	"runtime/metrics"
	"testing"
)

func readSchedHistogram() *metrics.Float64Histogram {
	sample := make([]metrics.Sample, 1)
	sample[0].Name = "/sched/latencies:seconds"
	metrics.Read(sample)
	return sample[0].Value.Float64Histogram()
}

func TestPercentiles(t *testing.T) {
	h := readSchedHistogram()
	v := Percentiles(h, []float64{0.5, 0.9, 0.99})
	t.Logf("p50: %f", v[0])
	t.Logf("p90: %f", v[1])
	t.Logf("p99: %f", v[2])
}
