package websocket

type Client interface {
	// return true to reconnect
	HandleDisconnect() bool
	HandleConnected()

	HandleMessage([]byte) error

	SendMessage([]byte)

	SendPing()

	Close()
}

type ChannelClient interface {
	Client

	SendEvent(event string, data interface{})

	SendSubscribe(channel string)
	SendUnsubscribe(channel string)

	Subscribe(channel string) *Channel
	Unsubscribe(channel string)
}

