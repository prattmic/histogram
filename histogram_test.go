package histogram

import (
	"os"
	"runtime/metrics"
	"sync"
	"testing"
	"time"
)

func readSchedHistogram() *metrics.Float64Histogram {
	sample := make([]metrics.Sample, 1)
	sample[0].Name = "/sched/latencies:seconds"
	metrics.Read(sample)
	return sample[0].Value.Float64Histogram()
}

func doSomeSchedWork() {
	ch := make(chan int)
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case v := <-ch:
					select {
					case <-done:
						return
					case ch <- v:
					}
				}
			}
		}()
	}
	ch <- 1
	time.Sleep(10*time.Millisecond)
	close(done)
	wg.Wait()
}

func TestPercentiles(t *testing.T) {
	doSomeSchedWork()

	h := readSchedHistogram()
	v := Percentiles(h, []float64{0.5, 0.9, 0.99})
	t.Logf("p50: %f", v[0])
	t.Logf("p90: %f", v[1])
	t.Logf("p99: %f", v[2])
}

func TestSamples(t *testing.T) {
	doSomeSchedWork()

	h := readSchedHistogram()
	s := Samples(h)
	t.Logf("samples: %d", s)
}

func TestVisualize(t *testing.T) {
	doSomeSchedWork()

	h := readSchedHistogram()
	s := Visualize(h, false)
	t.Logf(s)
}

func TestCSV(t *testing.T) {
	doSomeSchedWork()

	h := readSchedHistogram()
	if err := CSV(h, os.Stdout); err != nil {
		t.Errorf("CSV() got err %v want nil", err)
	}
}
