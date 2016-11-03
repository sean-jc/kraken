package kraken

import (
	"bytes"
	"time"
)

type Kraken struct {
	stopch    chan struct{}
	tentacles uint64
}

func New(tentacles uint64) (*Kraken, error) {
	k := &Kraken{
		stopch:    make(chan struct{}),
		tentacles: tentacles,
	}
	return k, nil
}

// Stop stops the current attack.
func (k *Kraken) Stop() {
	select {
	case <-k.stopch:
		return
	default:
		close(k.stopch)
	}
}

func (k *Kraken) hit(t *Target, tm time.Time) *Result {
	var err error

	res := Result{
		Timestamp: tm,
		Latencies: make([]Latency, len(t.post)+1),
	}
	res.Latencies[0].Name = "Overall"

	var b bytes.Buffer

	defer func() {
		res.Latencies[0].Time = time.Since(tm)

		if err == nil {
			err = t.ProcessOutput(&b, &res)
		}
		if err != nil {
			res.Error = err.Error()
		}
	}()

	err = t.Hit(&b)
	if err != nil {
		return &res
	}

	return &res
}
