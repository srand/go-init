package utils

import (
	"sync"
)

type Broker[T any] struct {
	subs  map[chan T]struct{}
	mutex sync.Mutex
}

func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		map[chan T]struct{}{},
		sync.Mutex{},
	}
}

func (b *Broker[T]) Subscribe() chan T {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	msgCh := make(chan T, 1)
	b.subs[msgCh] = struct{}{}
	return msgCh
}

func (b *Broker[T]) Unsubscribe(msgCh chan T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	delete(b.subs, msgCh)
}

func (b *Broker[T]) Publish(msg T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for msgCh := range b.subs {
		select {
		case msgCh <- msg:
		default:
		}
	}
}
