package kraken

import (
	"fmt"
	"strings"
	"time"
)

// Buckets represents an Histogram's latency buckets.
type Buckets []time.Duration

type histogramTable struct {
	Name   string
	Counts []uint64
	Total  uint64
}

// Histogram is a bucketed latency Histogram.
type Histogram struct {
	Buckets Buckets

	Tables []histogramTable
}

// Add implements the Add method of the Report interface by finding the right
// Bucket for the given Result latency and increasing its count by one as well
// as the total count.
func (h *Histogram) Add(r *Result) {
	h.init(r)

	for t, l := range r.Latencies {
		if len(h.Tables[t].Counts) != len(h.Buckets) {
			h.Tables[t].Counts = make([]uint64, len(h.Buckets))
		}

		var i int
		for ; i < len(h.Buckets)-1; i++ {
			if l.Time >= h.Buckets[i] && l.Time < h.Buckets[i+1] {
				break
			}
		}

		h.Tables[t].Name = l.Name
		h.Tables[t].Total++
		h.Tables[t].Counts[i]++
	}
}

// Nth returns the nth bucket represented as a string.
func (bs Buckets) Nth(i int) (left, right string) {
	if i >= len(bs)-1 {
		return bs[i].String(), "+Inf"
	}
	return bs[i].String(), bs[i+1].String()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (bs *Buckets) UnmarshalText(value []byte) error {
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return fmt.Errorf("bad buckets: %s", value)
	}
	for _, v := range strings.Split(string(value[1:len(value)-1]), ",") {
		d, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return err
		}
		*bs = append(*bs, d)
	}
	if len(*bs) == 0 {
		return fmt.Errorf("bad buckets: %s", value)
	}
	return nil
}

func (h *Histogram) init(r *Result) {
	if len(h.Tables) == 0 {
		h.Tables = make([]histogramTable, len(r.Latencies))
	}
}
