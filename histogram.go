// Package histogram provides helpers for working with runtime/metrics
// histograms.
package histogram

import (
	"io"
	"encoding/csv"
	"fmt"
	"runtime/metrics"
	"strings"
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
				lower := h.Buckets[i] * 1e9
				upper := h.Buckets[i+1] * 1e9
				mean := (lower + upper) / 2
				vals[j] = mean
			}
		}

	}

	return vals
}

func Percentile(h *metrics.Float64Histogram, pct float64) float64 {
	return Percentiles(h, []float64{pct})[0]
}

// Samples returns the total number of samples in the histogram.
func Samples(h *metrics.Float64Histogram) uint64 {
	var total uint64
	for _, count := range h.Counts {
		total += count
	}
	return total
}

// Visualize returns a rudimentary ASCII visualization. It plots buckets
// directly, despite differing bucket sizes, so the visualization may be
// misleading.
//
// If full, all buckets are shown. Otherwise the output is compressed to
// exclude most empty buckets.
func Visualize(h *metrics.Float64Histogram, full bool) string {
	var b strings.Builder

	var maxCount uint64
	for _, count := range h.Counts {
		if count > maxCount {
			maxCount = count
		}
	}

	shouldPrint := interestingBuckets(h, full)

	const maxWidth = 20
	for i, count := range h.Counts {
		if !shouldPrint[i] {
			continue
		} else if i > 0 && !shouldPrint[i-1] {
			// Didn't print last bucket, indicate skipped section.
			fmt.Fprintf(&b, "%20s| ...\n", " ")
		}

		lower := h.Buckets[i] * 1e9
		upper := h.Buckets[i+1] * 1e9

		width := 0
		if maxCount > 0 {
			width = int(maxWidth * (float64(count) / float64(maxCount)))
		}
		bar := strings.Repeat("*", width)

		fmt.Fprintf(&b, "%20s| %6d [%6.1f, %6.1f)\n", bar, count, lower, upper)
	}

	return b.String()
}

// interestingBuckets returns a slice of bool, where each index corresponds to
// whether that index of h.Count should be printed.
func interestingBuckets(h *metrics.Float64Histogram, full bool) []bool {
	interesting := make([]bool, len(h.Counts))
	if full {
		for i := range interesting {
			interesting[i] = true
		}
		return interesting
	}

	markSurrounding := func(i int) {
		const surround = 2 // 3 buckets around an interesting one are also interesting.

		// Buckets before i.
		for j := i-1; j >= i-surround && j >= 0; j-- {
			interesting[j] = true
		}

		// i itself.
		interesting[i] = true

		// Buckets after i.
		for j := i+1; j <= i+surround && j < len(interesting); j++ {
			interesting[j] = true
		}
	}

	// Start and end are always interesting.
	markSurrounding(0)
	markSurrounding(len(h.Counts)-1)

	for i, count := range h.Counts {
		if count > 0 {
			markSurrounding(i)
		}
	}

	return interesting
}

func CSV(h *metrics.Float64Histogram, w io.Writer) error {
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{"lower", "upper", "count"}); err != nil {
		return err
	}

	for i, count := range h.Counts {
		lower := fmt.Sprintf("%f", h.Buckets[i] * 1e9)
		upper := fmt.Sprintf("%f", h.Buckets[i+1] * 1e9)
		cnt := fmt.Sprintf("%d", count)

		if err := cw.Write([]string{lower, upper, cnt}); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
