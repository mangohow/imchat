package chatserver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/cmd/chatserver/internal/log"
	"github.com/mangohow/imchat/pkg/common/xwaitgroup"
	ws "github.com/mangohow/imchat/pkg/common/xwebsocket"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// 服务端监听的端口
	Addr string
	// 心跳时间
	HeartBeat time.Duration
}

/*
	websocket.Conn.WriteMessage方法不支持并发写操作，
	因此采用一个goroutine负责写，在其它地方调用的Write操作只会将数据放入chan中
	写goroutine从chan中读取并执行真正的写操作给客户端

	websocket.Conn.WriteControl方法中使用了一个chan来防止并发写操作，
	因此WriteControl方法可用并发写
*/

type ChatServer struct {
	// 消息处理器
	messageHandler IMessageHandler

	ws            *ws.WebSocket
	// 客户端管理器
	clientManager IClientManager

	nodeId string

	logger *logrus.Logger

	afterHandshakeHandler   AfterHandshakeHandler
	afterClientCloseHandler AfterClientCloseHandler

	config *Config

	ctx context.Context
	cancel context.CancelFunc

	waitgroup xwaitgroup.WaitGroupWrapper
}

type AfterHandshakeHandler func(r *http.Request, conn *Client) bool
type AfterClientCloseHandler func(conn *Client)

const DefaultListenAddr = "127.0.0.1:8000"

func NewServer(conf *Config) *ChatServer {
	if conf == nil {
		conf = &Config{
			Addr: DefaultListenAddr,
			HeartBeat: 0,
		}
	}

	if conf.Addr == "" {
		conf.Addr = DefaultListenAddr
	}

	// 生成serverID
	id := generateServerId()

	log.Logger().Debug("node id:", id)

	ctx, cancelFunc := context.WithCancel(context.Background())
	return &ChatServer{
		clientManager:  ClientManagerInstance,
		messageHandler: NewMessageHandler(),
		ws:             ws.New("/ws/chat"),
		logger:         log.Logger(),
		config:         conf,
		nodeId:         id,
		ctx:            ctx,
		cancel:         cancelFunc,
	}
}

// 生成serverID，并将ID保存到文件中，服务重启时使用原有的
func generateServerId() (id string) {
	var filename string
	if conf.ServerConf.NodeId != 0 {
		filename = fmt.Sprintf("chatserver%d.id", conf.ServerConf.NodeId)
	} else {
		filename = "chatserver.id"
	}

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		file, err = os.Create(filename)
		if err != nil {
			panic(err)
		}

		// 生成nodeId
		id = uuid.New().String()

		_, err = file.Write([]byte(id))
		if err != nil {
			panic(err)
		}
	} else {
		b, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		id = string(b)
	}
	file.Close()

	return
}

func (s *ChatServer) SetAfterHandshakeHandler(handler AfterHandshakeHandler) {
	s.afterHandshakeHandler = handler
}

// Serve 服务启动时需要注册到消息队列
func (s *ChatServer) Serve() error {
	s.ws.HandleWebSocket(s.websocketHandler)
	s.logger.Info("server listen at: ", s.config.Addr)
	return s.ws.ListenAndServe(s.config.Addr)
}

const MessageTypeLen = 4

// 采用读写分离的方式
func (s *ChatServer) startClientWriter(ctx context.Context, conn *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <- conn.ch:
			if isControl(res.wsMsgType) {
				conn.wsc.WriteControl(res.wsMsgType, res.data, time.Now().Add(s.config.HeartBeat))
			} else {
				conn.wsc.WriteMessage(res.wsMsgType, res.data)
			}

		}

	}
}

// websocketHandler 用户认证成功后, 需要添加到redis中
func (s *ChatServer) websocketHandler(conn *websocket.Conn, r *http.Request) {
	defer conn.Close()
	cli := NewClient(conn)
	// 调用握手后的处理方法
	if s.afterHandshakeHandler != nil {
		if !s.afterHandshakeHandler(r, cli) {
			return
		}
	}

	// 调用客户端被关闭后的handler
	if s.afterClientCloseHandler != nil {
		defer s.afterClientCloseHandler(cli)
	}

	// 设置心跳
	if s.config.HeartBeat > 0 {
		conn.SetPingHandler(func(_ string) error {
			conn.WriteControl(websocket.PongMessage, nil, time.Unix(0, 0))
			return conn.SetReadDeadline(time.Now().Add(s.config.HeartBeat))
		})
	}

	// 启动writer
	ctx, cancelFunc := context.WithCancel(s.ctx)
	defer cancelFunc()

	s.waitgroup.Go(func() {
		s.startClientWriter(ctx, cli)
	})

	id, _ := cli.GetInt64("id")
	s.logger.Debugf("[new connection] addr: %s, id: %d", conn.RemoteAddr(), id)

	for {
		err := conn.SetReadDeadline(time.Now().Add(s.config.HeartBeat))
		if err != nil {
			s.logger.Errorf("set heartbeat error:%v", err)
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				s.logger.Errorf("read timeout, ip:%v", conn.RemoteAddr().String())
				return
			}

			if e, ok := err.(*websocket.CloseError);ok {
				switch e.Code {
				case websocket.CloseNormalClosure:
					return
				case websocket.CloseAbnormalClosure:
					s.logger.Errorf("client shutdown abnormal")
					return
				}
			}

			s.logger.Errorf("read message error:%v", err)
			return
		}

		if len(data) < MessageTypeLen {
			s.logger.Errorf("data invalid")
			return
		}

		// 获取messageType
		msgType := binary.LittleEndian.Uint32(data[:MessageTypeLen])

		msg := &Message{
			MsgId:     msgType,
			RawData:   data,
			ProtoData: data[MessageTypeLen:],
		}

		err = s.handleRequest(cli, msg)
		if err != nil {
			s.logger.Errorf("handle request error:%v", err)
			return
		}
	}
}

func (s *ChatServer) HeartBeat() time.Duration {
	return s.config.HeartBeat
}

func (s *ChatServer) ServerId() string {
	return s.nodeId
}


func (s *ChatServer) SetClientCloseHandler(handler AfterClientCloseHandler) {
	s.afterClientCloseHandler = handler
}

func (s *ChatServer) handleRequest(conn *Client, message *Message) error {
	return s.messageHandler.Handle(conn, message)
}

func (s *ChatServer) Use(midFunc ...HandlerFunc) {
	s.messageHandler.Use(midFunc...)
}

func (s *ChatServer) HandlerFunc(id uint32, handler HandlerFunc) {
	s.messageHandler.HandlerFunc(id, handler)
}

// HandlerAnyFunc 传入的函数必须遵循下面规则：
// 一个入参：func(ctx *Context) proto.Message类型指针返回值 或 没有返回值
// 两个入参: func(ctx *Context, proto.Message类型的指针) *proto.Message类型指针返回值 或没有返回值
func (s *ChatServer) HandlerAnyFunc(id uint32, handler AnyFunc) {
	s.messageHandler.HandlerAnyFunc(id, handler)
}

func (s *ChatServer) Group(handlers ...HandlerFunc) {
	s.messageHandler.Group(handlers...)
}

func (s *ChatServer) GetCtx() context.Context {
	return s.ctx
}

func (s *ChatServer) Shutdown() error {
	s.cancel()
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	err := s.ws.Shutdown(ctx)

	s.waitgroup.Wait()

	return err
}

func (s *ChatServer) RegisterOnShutdown(fn func()) {
	s.ws.RegisterOnShutdown(fn)
}

func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}