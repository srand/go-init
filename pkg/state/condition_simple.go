package state

type ConditionCheckFunc func(update func(bool))

type SimpleCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
	check      ConditionCheckFunc
}

func (s *SimpleCond) Name() string {
	return s.name
}

func (s *SimpleCond) Get() bool {
	return s.value
}

func (s *SimpleCond) Set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *SimpleCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
	observer.OnChange(s, s.Get())
}

func (s *SimpleCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *SimpleCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewCondition(name string, check ConditionCheckFunc) *SimpleCond {
	cond := &SimpleCond{
		observable: NewObservable[Condition, bool](),
		name:       name,
		check:      check,
	}

	// cond.Subscribe(&ConditionLogger{})

	if check != nil {
		go check(cond.Set)
	}

	return cond
}
