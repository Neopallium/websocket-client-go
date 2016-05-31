package websocket

import (
	"encoding/json"
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
	in                 chan *Event
	out                chan []byte
	closeSocket        chan bool
	lastActivity       time.Time
	connectTimeout     time.Duration
	activityTimeout    time.Duration
	pingTimeout        time.Duration
	connectDelay       time.Duration
	timeoutTimer       *TimeoutTimer
}

func (s *Socket) setTimeout(reason TimeoutReason, d time.Duration) {
	s.timeoutTimer.SetTimeout(reason, d)
}

func (s *Socket) updateActivity() {
	s.timeoutTimer.Reset()
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
	s.setTimeout(PingTimeout, s.pingTimeout)
	// send ping
	s.SendMsg([]byte(`{"event":"pusher:ping","data":"{}"}`))
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
	s.setTimeout(ConnectTimeout, s.connectTimeout)
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

func (s *Socket) handleError(event *Event) stateFn {
	var msg struct {
		Message string
		Code    int
	}
	if err := json.Unmarshal([]byte(event.Data), &msg); err != nil {
		log.Println("Failed to unmarshal:", event.Event, err)
	}
	switch {
	case 4000 <= msg.Code && msg.Code <= 4099:
		log.Println("Connect failed websocket error: code:", msg.Code, ", message:", msg.Message)
		return stopState
	case 4100 <= msg.Code && msg.Code <= 4199:
		log.Println("Try again (delayed reconnect): code:", msg.Code, ", message:", msg.Message)
		s.connectDelay = time.Second
		return reconnectState
	case 4200 <= msg.Code && msg.Code <= 4299:
		log.Println("Reconnect (no delay): code:", msg.Code, ", message:", msg.Message)
		s.connectDelay = 0
		return reconnectState
	default:
		log.Println("Websocket error: code:", msg.Code, ", message:", msg.Message)
	}
	return connectedState
}

func (s *Socket) handleConnectionEstablished(event *Event) stateFn {
	// process Connected information.
	var msg struct {
		SocketId         string `json:"socket_id"`
		ActivityTimeout  int `json:"activity_timeout"`
	}
	if err := json.Unmarshal([]byte(event.Data), &msg); err != nil {
		log.Println("Failed to unmarshal:", event.Event, err)
	}
	// check activity_timeout value
	activityTimeout := time.Duration(msg.ActivityTimeout) * time.Second
	if(activityTimeout > 0 && activityTimeout < s.activityTimeout) {
		s.activityTimeout = activityTimeout
	}
	// set activity timeout
	s.setTimeout(ActivityTimeout, s.activityTimeout)
	// reset connect delay
	s.connectDelay = 0
	s.client.HandleConnected()
	return connectedState
}

func (s *Socket) messageState(event *Event) stateFn {
	next := connectedState
	// handle Websocket events.
	switch event.Event {
	case "pusher:ping":
		s.SendMsg([]byte(`{"event":"pusher:pong","data":"{}"}`))
	case "pusher:pong":
		// change from ping timeout to activity timeout
		s.setTimeout(ActivityTimeout, s.activityTimeout)
	case "pusher:error":
		next = s.handleError(event)
	case "pusher:connection_established":
		next = s.handleConnectionEstablished(event)
	}
	s.client.HandleMessage(event)
	return next
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
			return s.messageState(event)
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

