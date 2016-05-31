package websocket

import (
	"time"
	"net"
)

type DelayError interface {
	net.Error
	Delay() time.Duration  // Delay reconnect?
}

type Error struct {
	reason   string
	timeout  bool
	temp     bool
	delay    time.Duration
}

func (e *Error) Error() string {
	return e.reason
}

func (e *Error) Tiemout() bool {
	return e.timeout
}

func (e *Error) Temporary() bool {
	return e.temp
}

func (e *Error) Delay() time.Duration {
	return e.delay
}

func NewError(reason string, timeout bool, temp bool, delay time.Duration) *Error {
	return &Error{
		reason: reason,
		timeout: timeout,
		temp: temp,
		delay: delay,
	}
}

var (
	ErrReconnect = NewError("Reconnect (no delay)", false, true, 0)
	ErrDelayReconnect = NewError("Reconnect (with delay)", false, true, time.Second)
	ErrClosed = NewError("Closed", false, false, 0)
)

