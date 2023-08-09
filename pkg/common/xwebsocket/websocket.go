package xwebsocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	*http.Server
	WSPath string
}

type HandlerFunc func(conn *websocket.Conn, r *http.Request)


func New(wsPath string) *WebSocket {
	return &WebSocket{
		Server: &http.Server{},
		WSPath: wsPath,
	}
}

func (w *WebSocket) HandleWebSocket(handler HandlerFunc) {
	w.Server.Handler = http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if r.URL.Path != w.WSPath {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}

		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(writer, r, nil)
		if err != nil {
			return
		}


		handler(conn, r)
	})
}

func (w *WebSocket) ListenAndServe(addr string) error {
	w.Server.Addr = addr
	return w.Server.ListenAndServe()
}
