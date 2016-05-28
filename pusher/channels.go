package pusher

import (
	"sync"
)

type Channels struct {
	sync.RWMutex
	channels  map[string]*Channel
	global    *Channel
	connected bool
}

func (c *Channels) handleEvent(event *Event) {
	// send event to global channel
	c.global.handleEvent(event)
	if event.Channel == "" {
		// global only event.
		return
	}
	// send event to subscribed channel
	ch := c.find(event.Channel)
	if ch != nil {
		ch.handleEvent(event)
	}
}

func (c *Channels) connectedState(connected bool) {
	c.RLock()
	defer c.RUnlock()
	// cache connected state
	c.connected = connected
	// notify all channels of the connection state
	for _, ch := range c.channels {
		ch.connectedState(connected)
	}
}

func (c *Channels) find(channel string) *Channel {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		return c.global
	}
	// mutex is for the 'channels' map
	c.RLock()
	defer c.RUnlock()
	return c.channels[channel]
}

func (c *Channels) add(channel string, client *Client) *Channel {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		return c.global
	}
	// mutex is for the 'channels' map
	c.Lock()
	defer c.Unlock()
	// check if the channel exists.
	ch := c.channels[channel]
	if ch == nil {
		// create a new channel
		ch = newChannel(channel, client)
		c.channels[channel] = ch
	}
	if c.connected {
		ch.subscribe()
	}
	return ch
}

func (c *Channels) remove(channel string) {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		// can't remove the global channel.
		return
	}
	// mutex is for the 'channels' map
	c.Lock()
	defer c.Unlock()
	ch := c.channels[channel]
	if ch != nil && c.connected {
		ch.unsubscribe()
	}
	delete(c.channels, channel)
}

func (c *Channels) bind(event string, h Handler) {
	c.global.Bind(event, h)
}

func (c *Channels) unbind(event string, h Handler) {
	c.global.Unbind(event, h)
}

func newChannels(client *Client) *Channels {
	return &Channels{
		channels: make(map[string]*Channel),
		global: newChannel("", client),
	}
}

