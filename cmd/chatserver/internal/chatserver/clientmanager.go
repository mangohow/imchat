package chatserver

import (
	"sync"
)

type IClientManager interface {
	// Get 根据id获取Client
	Get(id int64) *Client
	
	// Add 添加一个Client
	Add(id int64, c *Client)
	
	// Del 删除一个Client
	Del(id int64)

	// Clear 清理所有连接，关闭连接并delete
	Clear()
}

var ClientManagerInstance = IClientManager(&clientManager{
	clients: make(map[int64]*Client),
})

type clientManager struct {
	clients map[int64]*Client
	rwm     sync.RWMutex
}


func (m *clientManager) Get(id int64) *Client {
	m.rwm.RLock()
	c := m.clients[id]
	m.rwm.RUnlock()
	return c
}

func (m *clientManager) Add(id int64, c *Client) {
	m.rwm.Lock()
	m.clients[id] = c
	m.rwm.Unlock()
}

func (m *clientManager) Del(id int64) {
	m.rwm.Lock()
	delete(m.clients, id)
	m.rwm.Unlock()
}

func (m *clientManager) Clear() {
	m.rwm.Lock()
	for _, client := range m.clients {
		_ = client.Close()
	}
	m.clients = make(map[int64]*Client)
	m.rwm.Unlock()
}