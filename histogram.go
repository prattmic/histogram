package histogram

import (
	"fmt"
	"runtime/metrics"
)

func Percentiles(h *metrics.Float64Histogram, pct []float64) []float64 {
	for _, p := range pct {
		if p < 0 || p > 1 {
			panic(fmt.Sprintf("invalid percentile %f", p))
		}
	}

	var totalCount uint64
	for _, count := range h.Counts {
		totalCount += count
	}

	// Default to max value in case this percentile falls in the
	// last bucket.
	maxVal := h.Buckets[len(h.Buckets)-1]
	vals := make([]float64, len(pct))
	for i := range vals {
		vals[i] = maxVal
	}

	var runningCount uint64
	for i, count := range h.Counts {
		runningCount += count

		currPercent := float64(runningCount) / float64(totalCount)

		for j, p := range pct {
			if vals[j] == maxVal && currPercent >= p {
				lowerNanos := h.Buckets[i] * 1e9
				upperNanos := h.Buckets[i+1] * 1e9
				mean := (lowerNanos + upperNanos) / 2
				vals[j] = mean
			}
		}

	}

	return vals
}

func Percentile(h *metrics.Float64Histogram, pct float64) float64 {
	return Percentiles(h, []float64{pct})[0]
}
