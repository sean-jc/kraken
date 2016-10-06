package kraken

import (
	"sync"
	"time"
)

func (k *Kraken) Pummel(t *Target, hits uint64) <-chan *Result {
	var workers sync.WaitGroup
	results := make(chan *Result)
	targets := make(chan *Target, hits)

	// Spawn a goroutine for tentacle.
	for i := uint64(0); i < k.tentacles; i++ {
		go k.pummel(&workers, targets, results)
	}

	// Fill the channel with the target, each goroutine will pull a target off the
	// channel when its
	for i := uint64(0); i < hits; i++ {
		targets <- t
	}
	close(targets)

	// We don't need to listen to the stop channel as we can simply wait for the
	// workers to finish as Pummel doesn't need to do any worker adjustment.
	go func() {
		defer close(results)
		workers.Wait()
	}()

	return results
}

func (k *Kraken) pummel(workers *sync.WaitGroup, targets <-chan *Target, results chan<- *Result) {
	workers.Add(1)
	defer workers.Done()

	for t := range targets {
		select {
		case <-k.stopch:
			return
		default:
			results <- k.hit(t, time.Now())
		}
	}
}
