package chatserver

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

type ProtoHandlerGroup struct {
	*messageHandler
	chain HandlerChain
}

func NewMessageHandler() IMessageHandler {
	return &ProtoHandlerGroup{
		messageHandler: newMessageHandler(),
	}
}

func (h *ProtoHandlerGroup) Handle(conn *Client, message *Message) error {
	return h.handle(conn, message)
}

func (h *ProtoHandlerGroup) Use(midFunc ...HandlerFunc) {
	if len(h.chain) == 0 {
		h.chain = make(HandlerChain, 0, len(midFunc) + 1)
	}

	if len(h.chain) + len(midFunc) > int(abortIndex) -1 {
		panic(fmt.Sprintf("Max handler num:%d, now is:%d\n", abortIndex- 1, len(midFunc)))
	}

	h.chain = append(h.chain, midFunc...)
}

func (h *ProtoHandlerGroup) HandlerFunc(id uint32, handler HandlerFunc) {
	h.chain = append(h.chain, handler)
	h.addHandlers(id, h.chain)
}

var (
	ctxPointerType = reflect.TypeOf(&Context{})
)

// HandlerAnyFunc 使用反射对参数进行绑定
func (h *ProtoHandlerGroup) HandlerAnyFunc(id uint32, handler AnyFunc) {
	fv := reflect.ValueOf(handler)
	ft := fv.Type()

	// 传入的handler必须为函数
	if ft.Kind() != reflect.Func {
		panic("handler must be func type")
	}

	// 检查入参
	if inCount := ft.NumIn(); inCount == 0 || inCount > 2 {
		panic("in parameter must have one or two")
	}

	// 检查返回值
	if ft.NumOut() > 2 {
		panic("return value must have no more than two")
	}

	// 检查入参类型
	for i := 0; i < ft.NumIn(); i++ {
		in := ft.In(i)
		if in.Kind() != reflect.Pointer {
			panic("in parameter must be pointer")
		}

		if in == ctxPointerType {
			continue
		}


		// 检查是否是proto.Message类型
		value := reflect.New(in.Elem()).Interface()
		if _, ok := value.(proto.Message); !ok {
			panic("must be proto.Message type")
		}
	}

	// 检查返回值类型
	if ft.NumOut() == 1 {
		out := ft.Out(0)
		if out.Kind() != reflect.Pointer {
			panic("return value must be pointer")
		}
		value := reflect.New(out.Elem())
		if _, ok := value.Interface().(proto.Message); !ok {
			panic("return value must be proto.Message type")
		}
	}

	fn := func(ctx *Context) {
		inValues := make([]reflect.Value, 0, ft.NumIn())
		for i := 0; i < ft.NumIn(); i++ {
			in := ft.In(i)
			// 如果当前类型为*gin.Context，则将ctx注入
			if in == ctxPointerType {
				inValues = append(inValues, reflect.ValueOf(ctx))
				continue
			}

			in = in.Elem()
			inVal := reflect.New(in)
			// 否则，序列化为protobuf类型数据
			data := ctx.Message.ProtoData
			err := proto.Unmarshal(data, inVal.Interface().(proto.Message))
			if err != nil {
				panic(err)
			}

			inValues = append(inValues, inVal)
		}

		outValues := fv.Call(inValues)
		if len(outValues) == 0 {
			return
		}
		result := outValues[0].Interface()
		if result == nil {
			return
		}

		if ctx.GetRespId() == 0 {
			panic("not set resp id")
		}

		err := ctx.WriteProtoMessage( ctx.GetRespId(), result.(proto.Message))
		if err != nil {
			panic(err)
		}
	}


	h.chain = append(h.chain, fn)
	h.addHandlers(id, h.chain)
}

func (h *ProtoHandlerGroup) combineRootGroup(handlers HandlerChain) {
	if len(h.chain) == 0 {
		h.chain = make(HandlerChain, 0, len(handlers))
	}

	if len(handlers) > int(abortIndex) - 1 {
		panic(fmt.Sprintf("Max handler num:%d, now is:%d\n", abortIndex- 1, len(handlers)))
	}

	h.chain = append(h.chain, handlers...)
}


func (h *ProtoHandlerGroup) Group(handlers ...HandlerFunc) IMessageHandler {
	g := &ProtoHandlerGroup{
		messageHandler: h.messageHandler,
	}

	g.combineRootGroup(h.chain)
	g.Use(handlers...)

	return g
}