package btcmarkets

import (
	"time"
)

// RateLimitValue represent a rate limiting value in the form of an int
type RateLimitValue int

var (
	rateLimit10 RateLimitValue = 10
	// rateLimit25 RateLimitValue = 25 // TODO
)

// rateLimit is a basic rate limiting struct based upon ideas from the official
// golang wiki.
type rateLimit struct {
	tick     *time.Ticker
	throttle chan *time.Time
}

// Start starts a rate limiter with a supplied rate limit (duration)
// and allowing a provided burst of actions.
//
// The burst actions are not available until they have built up, to prevent
// over-spending of available rate limiting space.
func (l *rateLimit) Start(rate time.Duration, burst RateLimitValue) error {
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
