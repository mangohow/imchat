package handlers

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/mangohow/imchat/cmd/chatserver/internal/chatserver"
	"github.com/mangohow/imchat/pkg/common/commutil"
	"github.com/mangohow/imchat/pkg/consts"
)

// RetryHandler 当给接收者发送了消息后，可能会因为消息丢失而导致接收者收不到消息
// 因此在转发数据后，将数据加入重试队列中，以便后续进行重试
// 但是重试并不会重发消息，而是告诉用户有新消息，让用户主动去拉取
// 如果收到接收者的ACK，就从中移除

type IRetryHandler interface {
	Add(userId int64)

	Remove(userId int64)
}


type RetryItem struct {
	id int64
	retried int8
	nextTrigger time.Time
}

type MinHeap struct {
	items []*RetryItem
}

func (m *MinHeap) Len() int {
	return len(m.items)
}

func (m *MinHeap) Less(i, j int) bool {
	return m.items[i].nextTrigger.Before(m.items[j].nextTrigger)
}

func (m *MinHeap) Swap(i, j int) {
	m.items[i], m.items[j] = m.items[j], m.items[i]
}

func (m *MinHeap) Push(x any) {
	m.items = append(m.items, x.(*RetryItem))
}

func (m *MinHeap) Pop() any {
	n := len(m.items)
	res := m.items[n-1]
	m.items = m.items[:n-1]
	return res
}


type RetryHandler struct {
	worker int
	maxRetry int8
	ctx context.Context

	items *commutil.ConcurrentSet[int64]

	hlock sync.Mutex
	heap *MinHeap

	nextItem *RetryItem

	ch chan int64

	cond sync.Cond
}

func NewRetryHandler(ctx context.Context, worker int, maxRetry int8) IRetryHandler {
	h := &MinHeap{items: make([]*RetryItem, 0)}
	heap.Init(h)
	handler := &RetryHandler{
		ctx: ctx,
		worker: worker,
		maxRetry: maxRetry,
		heap: h,
		items: commutil.NewConcurrentSet[int64](),
		ch: make(chan int64, 1024 * 1024),
	}

	handler.cond.L = &handler.hlock

	go handler.handleRetry()
	for i := 0; i < worker; i++ {
		go handler.retryWorker()
	}

	return handler
}

const (
	MinRetryDuration = time.Second
)

func (r *RetryHandler) Add(userId int64) {
	if r.items.IsSet(userId) {
		return
	}
	r.items.Set(userId)
	item := &RetryItem{
		id:          userId,
		retried:     0,
		nextTrigger: time.Now().Add(MinRetryDuration),
	}

	r.hlock.Lock()
	heap.Push(r.heap, item)
	r.hlock.Unlock()
	r.cond.Signal()
}

func (r *RetryHandler) Remove(userId int64) {
	r.items.Del(userId)
}

// 从最小堆中处理待重试的item
// 发送到ch中由retryWorker来通知客户端
func (r *RetryHandler) handleRetry() {
	s := time.Duration(0)
	for {
		select {
		case <- r.ctx.Done():
			return
		default:
			for {
				if r.nextItem != nil {
					if r.items.IsSet(r.nextItem.id) {
						r.retry(r.nextItem)
					}
					r.nextItem = nil
				}

				r.hlock.Lock()
				for r.heap.Len() == 0 {
					r.cond.Wait()
				}

				item := heap.Pop(r.heap).(*RetryItem)
				r.hlock.Unlock()
				// 如果已经被移除了
				if !r.items.IsSet(item.id) {
					continue
				}

				// 还没到触发时间
				if item.nextTrigger.After(time.Now()) {
					r.nextItem = item
					s = item.nextTrigger.Sub(time.Now())
					break
				}

				// 到达触发时间
				r.retry(item)
			}
			time.Sleep(s)
		}
	}
}

func (r *RetryHandler) retry(item *RetryItem) {
	r.ch <- item.id
	// 超过最大重试次数
	if item.retried + 1 >= r.maxRetry {
		r.items.Del(item.id)
		return
	}

	item.retried++
	item.nextTrigger = time.Now().Add(time.Second * time.Duration(item.retried))
	r.hlock.Lock()
	heap.Push(r.heap, item)
	r.hlock.Unlock()
}

// 发送重试消息的worker
func (r *RetryHandler) retryWorker() {
	for {
		select {
		case <- r.ctx.Done():
			return
		case id := <- r.ch:
			target := chatserver.ClientManagerInstance.Get(id)
			if target == nil {
				return
			}
			target.WriteData(consts.NewMessage, nil)
		}
	}
}

