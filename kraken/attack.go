package kraken

import (
	"sync"
	"time"
)

func (k *Kraken) Attack(t *Target, rate uint64, d time.Duration) <-chan *Result {
	var workers sync.WaitGroup
	results := make(chan *Result)
	ticks := make(chan time.Time)

	for i := uint64(0); i < k.tentacles; i++ {
		go k.attack(t, &workers, ticks, results)
	}

	go func() {
		defer close(results)
		defer workers.Wait()
		defer close(ticks)
		interval := 1e9 / rate
		hits := rate * uint64(d.Seconds())
		began, done := time.Now(), uint64(0)
		for {
			now, next := time.Now(), began.Add(time.Duration(done*interval))
			time.Sleep(next.Sub(now))
			select {
			case ticks <- max(next, now):
				if done++; done == hits {
					return
				}
			case <-k.stopch:
				return
			default: // all workers are blocked. start one more and try again
				go k.attack(t, &workers, ticks, results)
			}
		}
	}()

	return results
}

func (k *Kraken) attack(t *Target, workers *sync.WaitGroup, ticks <-chan time.Time, results chan<- *Result) {
	workers.Add(1)
	defer workers.Done()
	for tm := range ticks {
		select {
		case <-k.stopch:
			return
		default:
			results <- k.hit(t, tm)
		}
	}
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
