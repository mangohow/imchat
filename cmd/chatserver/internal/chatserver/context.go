package chatserver

import (
	"math"
	"sync"
)

type Message struct {
	MsgId     uint32
	RawData   []byte
	ProtoData []byte
}

type Context struct {
	*Client
	Message *Message

	chain      HandlerChain
	chainIndex uint8

	respId uint32      // 响应类型Id
}

var ctxPool = &sync.Pool{
	New: func() any {
		return &Context{}
	},
}

func newContext(cli *Client, message *Message) *Context {
	ctx := ctxPool.Get().(*Context)
	ctx.Client = cli
	ctx.Message = message

	return ctx
}

func freeContext(ctx *Context) {
	ctx.chain = nil
	ctx.chainIndex = 0
	ctx.Client = nil
	ctx.Message = nil
	ctx.respId = 0
	ctxPool.Put(ctx)
}

const abortIndex uint8 = math.MaxUint8/2 + 1

func (c *Context) Abort() {
	c.chainIndex = abortIndex + 1
}

func (c *Context) Next() {
	c.chainIndex++

	for c.chainIndex < uint8(len(c.chain)) {
		h := c.chain[c.chainIndex]
		h(c)
		c.chainIndex++
	}
}

func (c *Context) SetRespId(id uint32) {
	c.respId = id
}

func (c *Context) GetRespId() uint32 {
	return c.respId
}

