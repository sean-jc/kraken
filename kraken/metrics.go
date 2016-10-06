package kraken

import (
	"fmt"
	"time"

	"github.com/streadway/quantile"
)

type (
	// Metrics holds metrics computed out of a slice of Results which are used
	// in some of the Reporters
	Metrics struct {
		// Latencies holds computed request latency metrics.
		Latencies []LatencyMetrics `json:"latencies"`
		// First is the earliest timestamp in a Result set.
		Earliest time.Time `json:"earliest"`
		// Latest is the latest timestamp in a Result set.
		Latest time.Time `json:"latest"`
		// End is the latest timestamp in a Result set plus its latency.
		End time.Time `json:"end"`
		// Duration is the duration of the attack.
		Duration time.Duration `json:"duration"`
		// Wait is the extra time waiting for responses from targets.
		Wait time.Duration `json:"wait"`
		// Hits is the total number of hits on the target.
		Hits uint64 `json:"hits"`
		// Rate is the rate of hits per second.
		Rate float64 `json:"rate"`
		// Success is the percentage of non-error responses.
		Success float64 `json:"success"`
		// Errors is a set of unique errors returned by the targets during the attack.
		Errors []string `json:"errors"`

		errors    map[string]struct{}
		success   uint64
		latencies []*quantile.Estimator
	}

	// LatencyMetrics holds computed request latency metrics.
	LatencyMetrics struct {
		// Name describe what this latency metric measures
		Name string `json:"name"`
		// Total is the total latency sum of all requests in an attack.
		Total time.Duration `json:"total"`
		// Mean is the mean request latency.
		Mean time.Duration `json:"mean"`
		// P50 is the 50th percentile request latency.
		P50 time.Duration `json:"50th"`
		// P95 is the 95th percentile request latency.
		P95 time.Duration `json:"95th"`
		// P99 is the 99th percentile request latency.
		P99 time.Duration `json:"99th"`
		// Max is the maximum observed request latency.
		Max time.Duration `json:"max"`
	}
)

// Add implements the Add method of the Report interface by adding the given
// Result to Metrics.
func (m *Metrics) Add(r *Result) {
	m.init(r)

	m.Hits++

	if m.Earliest.IsZero() || m.Earliest.After(r.Timestamp) {
		m.Earliest = r.Timestamp
	}

	if r.Timestamp.After(m.Latest) {
		m.Latest = r.Timestamp
	}

	if end := r.End(); end.After(m.End) {
		m.End = end
	}

	for i, l := range r.Latencies {
		if m.Latencies[i].Name != l.Name {
			panic(fmt.Sprintf("Divergence in latency names: expected %s, actual %s\n", m.Latencies[i].Name, l.Name))
		}
		m.Latencies[i].Total += l.Time

		if l.Time > m.Latencies[i].Max {
			m.Latencies[i].Max = l.Time
		}

		m.latencies[i].Add(float64(l.Time))
	}

	if r.Error == "" {
		m.success++
	} else {
		if _, ok := m.errors[r.Error]; !ok {
			m.errors[r.Error] = struct{}{}
			m.Errors = append(m.Errors, r.Error)
		}
	}
}

// Close implements the Close method of the Report interface by computing
// derived summary metrics which don't need to be run on every Add call.
func (m *Metrics) Close() {
	m.init(nil)

	m.Duration = m.Latest.Sub(m.Earliest)
	m.Rate = float64(m.Hits) / m.Duration.Seconds()
	m.Wait = m.End.Sub(m.Latest)
	m.Success = float64(m.Hits-uint64(len(m.Errors))) / float64(m.Hits)

	for i := range m.Latencies {
		m.Latencies[i].Mean = time.Duration(float64(m.Latencies[i].Total) / float64(m.Hits))
		m.Latencies[i].P50 = time.Duration(m.latencies[i].Get(0.50))
		m.Latencies[i].P95 = time.Duration(m.latencies[i].Get(0.95))
		m.Latencies[i].P99 = time.Duration(m.latencies[i].Get(0.99))
	}
}

func (m *Metrics) init(r *Result) {
	if m.errors == nil {
		m.errors = map[string]struct{}{}
	}
	if r != nil {
		if len(m.latencies) == 0 {
			m.latencies = make([]*quantile.Estimator, len(r.Latencies))
			for i := range r.Latencies {
				m.latencies[i] = quantile.New(
					quantile.Known(0.50, 0.01),
					quantile.Known(0.95, 0.001),
					quantile.Known(0.99, 0.0005),
				)
			}
		}

		if len(m.Latencies) == 0 {
			m.Latencies = make([]LatencyMetrics, len(r.Latencies))
			for i, l := range r.Latencies {
				m.Latencies[i].Name = l.Name
			}
		}
	}
}
