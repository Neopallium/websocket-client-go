package websocket

import (
	"sync"
)

type Channels struct {
	sync.RWMutex
	client    ChannelClient
	channels  map[string]Channel
	global    Channel
	connected bool
}

func (c *Channels) HandleEvent(event Event) {
	// send event to global channel
	if c.global != nil {
		c.global.HandleEvent(event)
	}
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
		ch.UpdateClientState(connected)
	}
}

func (c *Channels) SubscriptionSucceded(channel string, succeded bool) {
	c.RLock()
	defer c.RUnlock()
	ch := c.Find(channel)
	if ch != nil {
		ch.SetActive(true)
	}
}

func (c *Channels) Find(channel string) Channel {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		return c.global
	}
	// mutex is for the 'channels' map
	c.RLock()
	defer c.RUnlock()
	return c.channels[channel]
}

func (c *Channels) Add(channel string, ch Channel) {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		c.global = ch
		return
	}
	// mutex is for the 'channels' map
	c.Lock()
	defer c.Unlock()
	c.channels[channel] = ch
	if c.connected {
		ch.Subscribe()
	}
}

func (c *Channels) Remove(channel string) {
	// empty channel name is for receiving events from all subscribed channels.
	if channel == "" {
		c.global = nil
	}
	// mutex is for the 'channels' map
	c.Lock()
	defer c.Unlock()
	ch := c.channels[channel]
	if ch != nil && c.connected {
		ch.Unsubscribe()
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
		channels: make(map[string]Channel),
	}
}

