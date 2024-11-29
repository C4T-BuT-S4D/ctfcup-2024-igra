package sync

import "context"

type CtxMutex struct {
	ch chan struct{}
}

func NewCtxMutex() *CtxMutex {
	return &CtxMutex{
		ch: make(chan struct{}, 1),
	}
}

func (mu *CtxMutex) Lock(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case mu.ch <- struct{}{}:
		return true
	}
}

func (mu *CtxMutex) Unlock() {
	<-mu.ch
}
