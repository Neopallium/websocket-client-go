package websocket

import (
	"sync"
)

type PublicChannel struct {
	sync.RWMutex
	channel    string
	client     ChannelClient
	handlers   map[string][]Handler
	active     bool
}

func (c *PublicChannel) HandleEvent(event Event) {
	c.RLock()
	defer c.RUnlock()
	// send event to callbacks bound to this event.
	for _, h := range c.handlers[event.Event] {
		h.HandleEvent(event)
	}
	// send to callbacks bound to all handlers.
	for _, h := range c.handlers[""] {
		h.HandleEvent(event)
	}
}

func (c *PublicChannel) UpdateClientState(connected bool) {
	c.RLock()
	defer c.RUnlock()
	if connected && ! c.active {
		// Client connected.  Make sure we subscribe to the chanenl.
		c.Subscribe()
	} else {
		// Client disconnect, de-activate the channel.
		c.active = false
	}
}

func (c *PublicChannel) SetActive(active bool) {
	c.RLock()
	defer c.RUnlock()
	c.active = active
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

func (c *PublicChannel) Bind(event string, h Handler) {
	c.Lock()
	defer c.Unlock()
	c.handlers[event] = append(c.handlers[event], h)
}

func (c *PublicChannel) Unbind(event string, h Handler) {
	c.Lock()
	defer c.Unlock()
	remove(c.handlers[event], h)
}

func (c *PublicChannel) BindAll(h Handler) {
	c.Bind("", h)
}

func (c *PublicChannel) UnbindAll(h Handler) {
	c.Unbind("", h)
}

func (c *PublicChannel) BindFunc(event string, h func(Event)) {
	c.Bind(event, HandlerFunc(h))
}

func (c *PublicChannel) UnbindFunc(event string, h func(Event)) {
	c.Unbind(event, HandlerFunc(h))
}

func (c *PublicChannel) BindAllFunc(h func(Event)) {
	c.Bind("", HandlerFunc(h))
}

func (c *PublicChannel) UnbindAllFunc(h func(Event)) {
	c.Unbind("", HandlerFunc(h))
}

func (c *PublicChannel) Subscribe() {
	if c.channel == "" {
		return
	}
	c.client.SendSubscribe(c.channel)
}

func (c *PublicChannel) Unsubscribe() {
	if c.channel == "" {
		return
	}
	c.client.SendUnsubscribe(c.channel)
}

func NewPublicChannel(channel string, client ChannelClient) *PublicChannel {
	c := &PublicChannel{
		channel: channel,
		client: client,
		handlers: make(map[string][]Handler),
	}
	return c
}

