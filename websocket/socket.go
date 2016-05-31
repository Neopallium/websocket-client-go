package websocket

import (
	"github.com/gorilla/websocket"
	"net/url"
	"log"
	"time"
)

const (
	IN_CHANNEL_SIZE = 100
	OUT_CHANNEL_SIZE = 10
)

type stateFn func(s *Socket) stateFn

const (
	MAX_RECONNECT_WAIT = time.Second * 30
)

type Socket struct {
	client             Client
	url                string
	ws                 *websocket.Conn
	in                 chan []byte
	out                chan []byte
	closeSocket        chan bool
	lastActivity       time.Time
	connectTimeout     time.Duration
	activityTimeout    time.Duration
	pingTimeout        time.Duration
	connectDelay       time.Duration
	timeoutTimer       *TimeoutTimer
}

func (s *Socket) SetTimeout(reason TimeoutReason, d time.Duration) {
	s.timeoutTimer.SetTimeout(reason, d)
}

func (s *Socket) updateActivity() {
	s.timeoutTimer.Reset()
}

func (s *Socket) SetActivityTimeout(activityTimeout time.Duration) {
	// set activity_timeout value
	if(activityTimeout > 0 && activityTimeout < s.activityTimeout) {
		s.activityTimeout = activityTimeout
	}
	// update activity timeout
	s.SetTimeout(ActivityTimeout, s.activityTimeout)
}

func (s *Socket) reset() {
	if s.out != nil {
		close(s.out)
		// create a new out channel
		s.out = make(chan []byte, OUT_CHANNEL_SIZE)
	}
	if s.ws != nil {
		s.ws.Close()
		s.ws = nil
	}
}

// Close Websocket and don't reconnect.
func (s *Socket) Close() {
	close(s.closeSocket)
}

func (s *Socket) sendPing() {
	// set ping timeout
	s.SetTimeout(PingTimeout, s.pingTimeout)
	// send ping
	s.client.SendPing()
}

func (s *Socket) HandlePong() {
	// change from ping timeout to activity timeout
	s.SetTimeout(ActivityTimeout, s.activityTimeout)
}

func startState(s *Socket) stateFn {
	// handle delayed re-connects
	if s.connectDelay > 0 {
		if s.connectDelay > MAX_RECONNECT_WAIT {
			s.connectDelay = MAX_RECONNECT_WAIT
		}
		time.Sleep(s.connectDelay)
	}
	// Set connection timeout on dialer
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = s.connectTimeout
	s.SetTimeout(ConnectTimeout, s.connectTimeout)
	ws, _, err := dialer.Dial(s.url, nil)
	if err != nil {
		log.Println("Error connecting:", err)
		// increase delay & reconnect
		s.connectDelay += time.Second
		return startState
	}
	// websocket connected
	s.ws = ws
	// Start reader & writer
	s.makeReader()
	s.makeWriter()
	s.client.HandleConnected()
	return connectedState
}

func reconnectState(s *Socket) stateFn {
	s.reset()
	if s.client.HandleDisconnect() {
		return startState
	}
	return nil
}

func stopState(s *Socket) stateFn {
	s.reset()
	return nil
}

func (s *Socket) HandleConnected() {
	s.connectDelay = 0
}

func (s *Socket) errorState(err error) stateFn {
	log.Println("Websocket error:", err)
	switch err := err.(type) {
	case DelayError:
		if err.Temporary() {
			delay := err.Delay()
			if delay > 0 {
				s.connectDelay += delay
			} else {
				s.connectDelay = 0
			}
			return reconnectState
		}
	}
	return stopState
}

func timeoutState(s *Socket) stateFn {
	switch s.timeoutTimer.Reason {
	case ActivityTimeout:
		s.sendPing()
	case ConnectTimeout:
		log.Println("Connect timeout.")
		return reconnectState
	case PingTimeout:
		log.Println("Ping timeout.")
		return reconnectState
	}
	return connectedState
}

func connectedState(s *Socket) stateFn {
	for {
		// wait for event from reader or heartbeat
		select {
		case event := <-s.in:
			if event == nil {
				return reconnectState
			}
			s.updateActivity()
			if err := s.client.HandleMessage(event); err != nil {
				return s.errorState(err)
			}
		case tick := <-s.timeoutTimer.C:
			// process timeout timer ticks
			if s.timeoutTimer.tickExpired(tick) {
				return timeoutState
			}
			continue
		case <-s.closeSocket:
			return stopState
		}
	}
	panic("not reached")
}

func (s *Socket) run() {
	for state := startState; state != nil; {
		state = state(s)
	}
}

func NewSocket(u *url.URL, cf Config, client Client) *Socket {
	s := &Socket{
		client: client,
		url: u.String(),
		connectTimeout: cf.ConnectTimeout,
		activityTimeout: cf.ActivityTimeout,
		pingTimeout: cf.PingTimeout,
		out: make(chan []byte, OUT_CHANNEL_SIZE),
		closeSocket: make(chan bool),
		timeoutTimer: newTimeoutTimer(NoTimeout, 0),
	}
	// start connect stat machine
	go s.run()
	return s
}

