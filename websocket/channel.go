package websocket

type Handler interface {
	HandleEvent(Event)
}

type HandlerFunc func(Event)

func (f HandlerFunc) HandleEvent(e Event) {
	f(e)
}

type Channel interface {
	Handler

	UpdateClientState(connected bool)

	SetActive(active bool)

	Subscribe()
	Unsubscribe()

	Bind(event string, h Handler)
	Unbind(event string, h Handler)

	BindAll(h Handler)
	UnbindAll(h Handler)

	BindFunc(event string, h func(Event))
	UnbindFunc(event string, h func(Event))

	BindAllFunc(h func(Event))
	UnbindAllFunc(h func(Event))

}

