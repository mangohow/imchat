package chatserver

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// Client websocket库不支持并发写，因此需要在写数据时进行加锁
// 采用读写分离的方式来解决: 启动两个goroutine来读写数据
type Client struct {
	wsc *websocket.Conn
	authed bool

	mux sync.RWMutex
	kvs map[string]interface{}

	ch chan writeData
}

type writeData struct {
	wsMsgType int
	data []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		wsc: conn,
		kvs: make(map[string]interface{}),
		ch: make(chan writeData, 1024),
	}
}

func (c *Client) Authed() bool {
	return c.authed
}

func (c *Client) SetAuthed() {
	c.authed = true
}

func (c *Client) Set(key string, val interface{}) {
	c.mux.Lock()
	c.kvs[key] = val
	c.mux.Unlock()
}

func (c *Client) Get(key string) (val interface{}, exist bool ) {
	c.mux.RLock()
	val, exist = c.kvs[key]
	c.mux.RUnlock()

	return
}

func (c *Client) GetString(key string) (string, bool) {
	val, exist := c.Get(key)
	if !exist {
		return "", false
	}
	return val.(string), true
}

func (c *Client) GetInt64(key string) (int64, bool) {
	val, exist := c.Get(key)
	if !exist {
		return 0, false
	}
	return val.(int64), true
}

func (c *Client) GetUid() int64 {
	id, _ := c.GetInt64("id")
	return id
}

func (c *Client) Write(data []byte) {
	c.ch <- writeData{
		wsMsgType: websocket.BinaryMessage,
		data:      data,
	}
}

func (c *Client) WriteData(respId uint32, data []byte) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, respId)
	if len(data) == 0 {
		c.Write(buf)
		return
	}
	buffer := bytes.NewBuffer(buf)
	buffer.Write(data)
	c.Write(buffer.Bytes())
}

func (c *Client) WriteProtoMessage(respId uint32, message proto.Message) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, respId)

	buffer := bytes.NewBuffer(buf)

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	buffer.Write(data)

	c.Write(buffer.Bytes())
	return nil
}

func (c *Client) WriteMessage(messageType int, data []byte) {
	c.ch <- writeData{
		wsMsgType: messageType,
		data:      data,
	}
}

func (c *Client) WriteControl(messageType int, data []byte, deadline time.Time) error {
	return c.wsc.WriteControl(messageType, data, deadline)
}

func (c *Client) Close() error {
	return c.wsc.Close()
}

func (c *Client) RemoteAddr() string {
	return c.wsc.RemoteAddr().String()
}
