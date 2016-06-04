package pusher

import (
	ws "github.com/Neopallium/websocket-client-go/websocket"
)

type PublicChannel struct {
	ws.PublicChannel
}

func (c *PublicChannel) HandleEvent(event ws.Event) {
	c.RLock()
	defer c.RUnlock()
	// mark channel as subscribed
	if event.GetEvent() == "pusher_internal:subscription_succeeded" {
		c.SetActive(true)
	}
	c.PublicChannel.HandleEvent(event)
}

func NewPublicChannel(channel string, client ws.ChannelClient) *PublicChannel {
	return &PublicChannel{
		PublicChannel: *ws.NewPublicChannel(channel, client),
	}
}

