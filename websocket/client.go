package websocket

type ChangeState int

const (
	NoChangeState ChangeState = iota
	ReconnectState
	CloseState
)

type Message interface {
	EncodeMessage() ([]byte, error)
	DecodeMessage([]byte) error
}

type Client interface {
	// return true to reconnect
	HandleDisconnect() bool
	HandleConnected()

	HandleEvent(Event) ChangeState

	SendEvent(event string, data interface{})

	Close()

	//HandleMessage(Message) StateChange

	//DecodeMessage([]byte) Message
}

type ChannelClient interface {
	Client

	Subscribe(channel string) *Channel
	Unsubscribe(channel string)
}

