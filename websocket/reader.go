package websocket

// reader goroutine
func (s *Socket) makeReader() {
	ws := s.ws
	in := make(chan []byte, IN_CHANNEL_SIZE)
	go func () {
		for {
			_, buf, err := ws.ReadMessage()
			if err != nil {
				// Close channel to signal that the WebSocket connection has closed.
				close(in)
				return
			}
			in <-buf
		}
	} ()
	s.in = in
}

