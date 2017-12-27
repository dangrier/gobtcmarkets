package btcmarkets

import (
	"time"
)

// rateLimit is a basic rate limiting struct based upon ideas from the official
// golang wiki.
type rateLimit struct {
	tick     *time.Ticker
	throttle chan *time.Time
}

// Start creates starts a rate limiter with a supplied rate limit (duration)
// and allowing a provided burst of actions.
func (l *rateLimit) Start(rate time.Duration, burst int) error {
	(*l).tick = time.NewTicker(rate)

	(*l).throttle = make(chan *time.Time, burst)

	go func() {
		for t := range l.tick.C {
			select {
			case (*l).throttle <- &t:
			default:
			}
		}
	}()

	return nil
}

// Limit performs the actual limiting behaviour according to the Start()
// parameters.
func (l *rateLimit) Limit() error {
	<-l.throttle
	return nil
}
