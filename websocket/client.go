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

type TextMessage string

func (t TextMessage) EncodeMessage() ([]byte, error) {
	return []byte(t), nil
}

func (t TextMessage) DecodeMessage(buf []byte) error {
	t = TextMessage(buf)
	return nil
}

func (t TextMessage) String() string {
	return string(t)
}

type Client interface {
	// return true to reconnect
	HandleDisconnect() bool
	HandleConnected()

	DecodeMessage([]byte) (Message, error)
	HandleMessage(Message) ChangeState

	SendMessage(Message)

	Close()
}

type ChannelClient interface {
	Client

	SendEvent(event string, data interface{})

	Subscribe(channel string) *Channel
	Unsubscribe(channel string)
}

