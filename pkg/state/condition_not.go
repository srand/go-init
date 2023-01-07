package state

type notCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	cond       Condition
	value      bool
}

func (s *notCond) Name() string {
	return "not " + s.cond.Name()
}

func (s *notCond) Get() bool {
	return s.value
}

func (s *notCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *notCond) OnChange(cond Condition, value bool) {
	s.set(!value)
}

func (s *notCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
	observer.OnChange(s, s.Get())
}

func (s *notCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *notCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewNotCondition(cond Condition) *notCond {
	not := &notCond{
		observable: NewObservable[Condition, bool](),
		cond:       cond,
		value:      !cond.Get(),
	}

	cond.Subscribe(not)
	not.Subscribe(&ConditionLogger{})

	return not
}
