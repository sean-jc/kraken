package kraken

import (
	"sync"
	"time"
)

func (k *Kraken) Siege(t *Target, d time.Duration) <-chan *Result {
	var workers sync.WaitGroup
	results := make(chan *Result)
	buffered := make(chan *Result, k.tentacles)
	targets := make(chan *Target, k.tentacles)

	// Spawn a goroutine for each tentacle.  Siege uses pummel and simply
	// buffers the results, sending another target down the channel when
	// a result is received before forwarding the result on to the actual
	// results channel.
	for i := uint64(0); i < k.tentacles; i++ {
		go k.pummel(&workers, targets, buffered)
	}

	go func() {
		defer close(results)
		defer close(buffered)
		defer workers.Wait()
		defer close(targets)

		for {
			select {
			case <-k.stopch:
				return
			case res := <-buffered:
				targets <- t
				results <- res
			}
		}
	}()

	if d != 0 {
		go func() {
			time.Sleep(d)
			k.Stop()
		}()
	}

	return results
}
