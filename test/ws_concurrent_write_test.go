package test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWSConcurrentWrite(t *testing.T) {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial("ws://127.0.0.1:6387/ws/chat", map[string][]string{
		"authorization": {"123456"},
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		go func() {
			for {
				// conn.WriteMessage(websocket.TextMessage, []byte("test"))
				conn.WriteControl(websocket.PingMessage, []byte("test"), time.Time{})
				time.Sleep(time.Millisecond * 10)
			}

		}()
	}

	select {

	}
}
