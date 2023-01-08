package state

import "sync"

type Observable[Object any, Value any] interface {
	Subscribe(observer Observer[Object, Value])
	Unsubscribe(observer Observer[Object, Value])
	Publish()
}

type SimpleObservable[Object any, Value any] struct {
	Observable[Object, Value]
	mutex     sync.Mutex
	observers map[Observer[Object, Value]]struct{}
}

func (o *SimpleObservable[Object, Value]) Subscribe(observer Observer[Object, Value]) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.observers[observer] = struct{}{}
}

func (o *SimpleObservable[Object, Value]) Unsubscribe(observer Observer[Object, Value]) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	delete(o.observers, observer)
}

func (o *SimpleObservable[Object, Value]) Publish(obj Object, val Value) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	for observer := range o.observers {
		observer.OnChange(obj, val)
	}
}

func NewObservable[Object any, Value any]() *SimpleObservable[Object, Value] {
	return &SimpleObservable[Object, Value]{
		observers: map[Observer[Object, Value]]struct{}{},
	}
}

type Observer[Object any, Value any] interface {
	OnChange(obj Object, value Value)
}
