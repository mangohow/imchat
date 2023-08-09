package xwaitgroup

import "sync"

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Go(fun func()) {
	w.Add(1)
	go func() {
		fun()
		w.Done()
	}()
}
