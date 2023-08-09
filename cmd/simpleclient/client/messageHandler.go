package client

type HandlerFunc func(data []byte)

type MessageHandler struct {
	handlers map[uint32]HandlerFunc
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{handlers: make(map[uint32]HandlerFunc)}
}

func (m *MessageHandler) Register(id uint32, handler HandlerFunc) {
	m.handlers[id] = handler
}

func (m *MessageHandler) Handle(id uint32, data []byte) {
	fn := m.handlers[id]
	if fn != nil {
		fn(data)
	}
}
