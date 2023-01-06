package state

type State[Value string] interface {
	Observable[State[Value], Value]
	Name() string
	Get() Value
	Set(value Value)
}

type StateObserver[Value string] Observer[State[Value], Value]
type StateGetFunc[Value string] func() Value
type StateSetFunc[Value string] func(value Value)

type simpleState struct {
	State[string]
	observable *SimpleObservable[State[string], string]
	name       string
	value      string
}

func (s *simpleState) Name() string {
	return s.name
}

func (s *simpleState) Get() string {
	return s.value
}

func (s *simpleState) Set(value string) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *simpleState) Subscribe(observer Observer[State[string], string]) {
	s.observable.Subscribe(observer)
}

func (s *simpleState) Unsubscribe(observer Observer[State[string], string]) {
	s.observable.Unsubscribe(observer)
}

func (s *simpleState) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewState(name string, value string) State[string] {
	state := &simpleState{
		observable: NewObservable[State[string], string](),
		name:       name,
		value:      value,
	}

	// state.Subscribe(&StateLogger{})

	return state
}
