package state

type ConditionCheckFunc func(update func(bool))

type simpleCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
	check      ConditionCheckFunc
}

func (s *simpleCond) Name() string {
	return s.name
}

func (s *simpleCond) Get() bool {
	return s.value
}

func (s *simpleCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *simpleCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
	observer.OnChange(s, s.Get())
}

func (s *simpleCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *simpleCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewCondition(name string, check ConditionCheckFunc) Condition {
	cond := &simpleCond{
		observable: NewObservable[Condition, bool](),
		name:       name,
		check:      check,
	}

	// cond.Subscribe(&ConditionLogger{})

	go check(cond.set)

	return cond
}
