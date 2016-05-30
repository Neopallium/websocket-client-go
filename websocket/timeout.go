package websocket

import (
	"time"
)

type TimeoutReason int

const (
	NoTimeout TimeoutReason = iota
	ConnectTimeout
	ActivityTimeout
	PingTimeout
)

type TimeoutTimer struct {
	C        <-chan time.Time
	Reason   TimeoutReason
	ticker   *time.Ticker
	duration time.Duration
	start    time.Time
}

func (t *TimeoutTimer) tickExpired(tick time.Time) bool {
	// check if timeout is disabled.
	if t.Reason == NoTimeout {
		// ignore tick
		return false
	}
	// check if the timeout has expired
	expire := tick.Sub(t.start)
	if expire >= t.duration {
		return true
	}
	return false
}

func (t *TimeoutTimer) SetTimeout(reason TimeoutReason, d time.Duration) {
	t.Reason = reason
	t.duration = d
	t.Reset()
}

func (t *TimeoutTimer) Reset() {
	t.start = time.Now()
}

func (t *TimeoutTimer) Expired() bool {
	// process all buffered ticks from ticker.
	for {
		select {
		case tick := <-t.C:
			if t.tickExpired(tick) {
				return true
			}
		default:
			// no more ticks
			return false
		}
	}
	return false
}

func (t *TimeoutTimer) Stop() {
	t.ticker.Stop()
}

func newTimeoutTimer(reason TimeoutReason, d time.Duration) *TimeoutTimer {
	t := &TimeoutTimer{
		Reason: reason,
		ticker: time.NewTicker(time.Second),
		duration: d,
		start: time.Now(),
	}
	t.C = t.ticker.C
	return t
}

