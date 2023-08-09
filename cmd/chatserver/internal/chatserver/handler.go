package chatserver

import (
	"errors"
	"fmt"
)

type HandlerFunc func(ctx *Context)

type HandlerChain []HandlerFunc

type AnyFunc any

type IMessageHandler interface {
	Handle(conn *Client, message *Message) error

	Use(midFunc ...HandlerFunc)

	HandlerFunc(id uint32, fn HandlerFunc)

	HandlerAnyFunc(id uint32, fn AnyFunc)

	Group(handlers ...HandlerFunc) IMessageHandler
}


type messageHandler struct {
	handlers map[uint32]HandlerChain
}

func newMessageHandler() *messageHandler {
	return &messageHandler{
		handlers: make(map[uint32]HandlerChain),
	}
}

var NoSuchHandlersError = errors.New("no such handler")

func (h *messageHandler) handle(conn *Client, message *Message) error {
	id := message.MsgId
	handlers, ok := h.handlers[id]
	if !ok || len(handlers) == 0 {
		return NoSuchHandlersError
	}

	ctx := newContext(conn, message)
	ctx.chain = handlers
	chainLen := uint8(len(ctx.chain))

	for ; ctx.chainIndex < chainLen; ctx.chainIndex++ {
		handler := ctx.chain[ctx.chainIndex]
		if handler != nil {
			handler(ctx)
		}
	}

	freeContext(ctx)

	return nil
}

func (h *messageHandler) addHandlers(id uint32, handlers HandlerChain) {
	if _, ok := h.handlers[id]; ok {
		panic(fmt.Sprintf("duplicate handler for %d", id))
	}
	h.handlers[id] = handlers
}
