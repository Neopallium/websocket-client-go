package websocket

import (
	"sync"
)

type Channels struct {
	sync.RWMutex
	client    ChannelClient
	channels  map[string]*Channel
	global    *Channel
	connected bool
}

func (c *Channels) HandleEvent(event Event) {
	// send event to global channel
	c.global.HandleEvent(event)
	if event.Channel == "" {
		// global only event.
		return
	}
	// send event to subscribed channel
	ch := c.Find(event.Channel)
	if ch != nil {
		ch.HandleEvent(event)
	}
}

func (c *Channels) ConnectedState(connected bool) {
	c.RLock()
	defer c.RUnlock()
	// cache connected state
	c.connected = connected
	// notify all channels of the connection state
	for _, ch := range c.channels {
		ch.connectedState(connected)
	}
}

func (c *Channels) Find(channel string) *Channel {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		return c.global
	}
	// mutex is for the 'channels' map
	c.RLock()
	defer c.RUnlock()
	return c.channels[channel]
}

func (c *Channels) Add(channel string) *Channel {
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
		ch = newChannel(channel, c.client)
		c.channels[channel] = ch
	}
	if c.connected {
		ch.subscribe()
	}
	return ch
}

func (c *Channels) Remove(channel string) {
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

func (c *Channels) Bind(event string, h Handler) {
	c.global.Bind(event, h)
}

func (c *Channels) Unbind(event string, h Handler) {
	c.global.Unbind(event, h)
}

func NewChannels(client ChannelClient) *Channels {
	return &Channels{
		client: client,
		channels: make(map[string]*Channel),
		global: newChannel("", client),
	}
}

