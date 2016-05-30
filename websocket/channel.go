package websocket

import (
	"sync"
)

type Handler interface {
	HandleEvent(Event)
}

type HandlerFunc func(Event)

func (f HandlerFunc) HandleEvent(e Event) {
	f(e)
}

type Channel struct {
	sync.RWMutex
	channel    string
	client     ChannelClient
	handlers   map[string][]Handler
	subscribed bool
}

func (c *Channel) HandleEvent(event Event) {
	c.RLock()
	defer c.RUnlock()
	// mark channel as subscribed
	if event.Event == "pusher_internal:subscription_succeeded" {
		c.subscribed = true
	}
	// send event to callbacks bound to this event.
	for _, h := range c.handlers[event.Event] {
		h.HandleEvent(event)
	}
	// send to callbacks bound to all handlers.
	for _, h := range c.handlers[""] {
		h.HandleEvent(event)
	}
}

func (c *Channel) connectedState(connected bool) {
	c.RLock()
	defer c.RUnlock()
	if connected && ! c.subscribed {
		c.subscribe()
	} else {
		c.subscribed = false
	}
}

func remove(a []Handler, h Handler) {
	n := len(a)
	if n == 0 {
		return
	}
	for i, f := range a {
		if f == h {
			// replace with last handler
			n--
			a[i] = a[n]
			a[n] = nil
		}
	}
	// trim array
	a = a[:n]
}

func (c *Channel) Bind(event string, h Handler) {
	c.Lock()
	defer c.Unlock()
	c.handlers[event] = append(c.handlers[event], h)
}

func (c *Channel) Unbind(event string, h Handler) {
	c.Lock()
	defer c.Unlock()
	remove(c.handlers[event], h)
}

func (c *Channel) BindAll(h Handler) {
	c.Bind("", h)
}

func (c *Channel) UnbindAll(h Handler) {
	c.Unbind("", h)
}

func (c *Channel) BindFunc(event string, h func(Event)) {
	c.Bind(event, HandlerFunc(h))
}

func (c *Channel) UnbindFunc(event string, h func(Event)) {
	c.Unbind(event, HandlerFunc(h))
}

func (c *Channel) BindAllFunc(h func(Event)) {
	c.Bind("", HandlerFunc(h))
}

func (c *Channel) UnbindAllFunc(h func(Event)) {
	c.Unbind("", HandlerFunc(h))
}

type subData struct {
	Channel     string `json:"channel"`
	Auth        string `json:"auth,omitempty"`
	ChannelData string `json:"channel_data,omitempty"`
}

func (c *Channel) subscribe() {
	if c.channel == "" {
		return
	}
	c.client.SendEvent("pusher:subscribe", &subData{
		Channel: c.channel,
	})
}

type unsubData struct {
	Channel     string `json:"channel"`
}

func (c *Channel) unsubscribe() {
	if c.channel == "" {
		return
	}
	c.client.SendEvent("pusher:unsubscribe", &unsubData{
		Channel: c.channel,
	})
}

func newChannel(channel string, client ChannelClient) *Channel {
	c := &Channel{
		channel: channel,
		client: client,
		handlers: make(map[string][]Handler),
	}
	return c
}

