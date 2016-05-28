package pusher

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
	"log"
	"time"
)

const (
	IN_CHANNEL_SIZE = 100
	OUT_CHANNEL_SIZE = 10
)

type Event struct {
	Event    string `json:"event,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Data     string `json:"data,omitempty"`
}

type stateFn func(c *Client) stateFn

const (
	MAX_RECONNECT_WAIT = time.Second * 30
)

type Client struct {
	url                string
	ws                 *websocket.Conn
	in                 chan *Event
	out                chan []byte
	channels           *Channels
	lastActivity       time.Time
	connectTimeout     time.Duration
	activityTimeout    time.Duration
	pingTimeout        time.Duration
	connectDelay       time.Duration
	timeoutTimer       *TimeoutTimer
}

func (c *Client) setTimeout(reason TimeoutReason, d time.Duration) {
	c.timeoutTimer.SetTimeout(reason, d)
}

func (c *Client) updateActivity() {
	c.timeoutTimer.Reset()
}

func (c *Client) reset() {
	c.channels.connectedState(false)
	if c.out != nil {
		close(c.out)
		// create a new out channel
		c.out = make(chan []byte, OUT_CHANNEL_SIZE)
	}
	if c.ws != nil {
		c.ws.Close()
		c.ws = nil
	}
}

func (c *Client) sendPing() {
	// set ping timeout
	c.setTimeout(PingTimeout, c.pingTimeout)
	// send ping
	c.sendMsg(`{"event":"pusher:ping","data":"{}"}`)
}

func startState(c *Client) stateFn {
	// handle delayed re-connects
	if c.connectDelay > 0 {
		if c.connectDelay > MAX_RECONNECT_WAIT {
			c.connectDelay = MAX_RECONNECT_WAIT
		}
		time.Sleep(c.connectDelay)
	}
	// Set connection timeout on dialer
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.connectTimeout
	c.setTimeout(ConnectTimeout, c.connectTimeout)
	ws, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		log.Println("Error connecting:", err)
		// increase delay & reconnect
		c.connectDelay += time.Second
		return startState
	}
	// websocket connected
	c.ws = ws
	// start reader
	c.makeReader()
	return connectedState
}

func reconnectState(c *Client) stateFn {
	c.reset()
	return startState
}

func stopState(c *Client) stateFn {
	c.reset()
	return nil
}

func (c *Client) handleError(event *Event) stateFn {
	var msg struct {
		Message string
		Code    int
	}
	if err := json.Unmarshal([]byte(event.Data), &msg); err != nil {
		log.Println("Failed to unmarshal:", event.Event, err)
	}
	switch {
	case 4000 <= msg.Code && msg.Code <= 4099:
		log.Println("Connect failed pusher error: code:", msg.Code, ", message:", msg.Message)
		return stopState
	case 4100 <= msg.Code && msg.Code <= 4199:
		log.Println("Try again (delayed reconnect): code:", msg.Code, ", message:", msg.Message)
		c.connectDelay = time.Second
		return reconnectState
	case 4200 <= msg.Code && msg.Code <= 4299:
		log.Println("Reconnect (no delay): code:", msg.Code, ", message:", msg.Message)
		c.connectDelay = 0
		return reconnectState
	default:
		log.Println("Pusher error: code:", msg.Code, ", message:", msg.Message)
	}
	return connectedState
}

func (c *Client) handleConnectionEstablished(event *Event) stateFn {
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
	if(activityTimeout > 0 && activityTimeout < c.activityTimeout) {
		c.activityTimeout = activityTimeout
	}
	// set activity timeout
	c.setTimeout(ActivityTimeout, c.activityTimeout)
	// reset connect delay
	c.connectDelay = 0
	// Start writer
	c.makeWriter()
	c.channels.connectedState(true)
	return connectedState
}

func (c *Client) handleEvent(event *Event) stateFn {
	next := connectedState
	// handle Pusher events.
	switch event.Event {
	case "pusher:ping":
		c.sendMsg(`{"event":"pusher:pong","data":"{}"}`)
	case "pusher:pong":
		// change from ping timeout to activity timeout
		c.setTimeout(ActivityTimeout, c.activityTimeout)
	case "pusher:error":
		next = c.handleError(event)
	case "pusher:connection_established":
		next = c.handleConnectionEstablished(event)
	}
	c.channels.handleEvent(event)
	return next
}

func timeoutState(c *Client) stateFn {
	switch c.timeoutTimer.Reason {
	case ActivityTimeout:
		c.sendPing()
	case ConnectTimeout:
		log.Println("Connect timeout.")
		return reconnectState
	case PingTimeout:
		log.Println("Ping timeout.")
		return reconnectState
	}
	return connectedState
}

func connectedState(c *Client) stateFn {
	for {
		// wait for event from reader or heartbeat
		select {
		case event := <-c.in:
			if event == nil {
				return reconnectState
			}
			c.updateActivity()
			return c.handleEvent(event)
		case tick := <-c.timeoutTimer.C:
			// process timeout timer ticks
			if c.timeoutTimer.tickExpired(tick) {
				return timeoutState
			}
			continue
		}
	}
	panic("not reached")
}

func (c *Client) run() {
	for state := startState; state != nil; {
		state = state(c)
	}
}

func (c *Client) Subscribe(channel string) *Channel {
	return c.channels.add(channel, c)
}

func (c *Client) Unsubscribe(channel string) {
	c.channels.remove(channel)
}

func (c *Client) Bind(event string, h Handler) {
	c.channels.bind(event, h)
}

func (c *Client) Unbind(event string, h Handler) {
	c.channels.unbind(event, h)
}

func (c *Client) BindFunc(event string, h func(Event)) {
	c.Bind(event, HandlerFunc(h))
}

func (c *Client) UnbindFunc(event string, h func(Event)) {
	c.Unbind(event, HandlerFunc(h))
}

func (c *Client) BindAll(h Handler) {
	c.Bind("", h)
}

func (c *Client) UnbindAll(h Handler) {
	c.Unbind("", h)
}

func (c *Client) BindAllFunc(h func(Event)) {
	c.Bind("", HandlerFunc(h))
}

func (c *Client) UnbindAllFunc(h func(Event)) {
	c.Unbind("", HandlerFunc(h))
}

func newClient(u *url.URL, p *Pusher) *Client {
	params := u.Query()
	params.Set("protocol", strconv.Itoa(p.Protocol))
	params.Set("version", p.Version)
	params.Set("client", p.Client)
	u.RawQuery = params.Encode()
	c := &Client{
		url: u.String(),
		connectTimeout: p.ConnectTimeout,
		activityTimeout: p.ActivityTimeout,
		pingTimeout: p.PingTimeout,
		out: make(chan []byte, OUT_CHANNEL_SIZE),
		timeoutTimer: newTimeoutTimer(NoTimeout, 0),
	}
	c.channels = newChannels(c)
	// start connect stat machine
	go c.run()
	return c
}

