package state

type State[Value string | int] interface {
	Observable[State[Value], Value]
	Name() string
	Get() Value
	Set(value Value)
}

type StateObserver[Value string] Observer[State[Value], Value]
type StateGetFunc[Value string] func() Value
type StateSetFunc[Value string] func(value Value)

type simpleState[T string | int] struct {
	State[string]
	observable *SimpleObservable[State[T], T]
	name       string
	value      T
}

func (s *simpleState[T]) Name() string {
	return s.name
}

func (s *simpleState[T]) Get() T {
	return s.value
}

func (s *simpleState[T]) Set(value T) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *simpleState[T]) Subscribe(observer Observer[State[T], T]) {
	s.observable.Subscribe(observer)
}

func (s *simpleState[T]) Unsubscribe(observer Observer[State[T], T]) {
	s.observable.Unsubscribe(observer)
}

func (s *simpleState[T]) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewState[T string | int](name string, value T) State[T] {
	state := &simpleState[T]{
		observable: NewObservable[State[T], T](),
		name:       name,
		value:      value,
	}

	state.Subscribe(&StateLogger[T]{})

	return state
}
