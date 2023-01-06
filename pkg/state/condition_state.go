package state

type stateCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
	state      State[string]
	stateValue string
}

func (s *stateCond) Name() string {
	return s.name
}

func (s *stateCond) Get() bool {
	return s.value
}

func (s *stateCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *stateCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
}

func (s *stateCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *stateCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func (s *stateCond) OnChange(state State[string], value string) {
	s.set(s.stateValue == value)
}

func NewStateCondition(name string, state State[string], value string) Condition {
	cond := &stateCond{
		observable: NewObservable[Condition, bool](),
		name:       name,
		state:      state,
		stateValue: value,
		value:      state.Get() == value,
	}

	state.Subscribe(cond)
	cond.Subscribe(&ConditionLogger{})

	return cond
}
