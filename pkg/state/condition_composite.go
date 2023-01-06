package state

type compositeCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
	conditions []Condition
}

func (s *compositeCond) AddCondition(cond Condition) {
	s.conditions = append(s.conditions, cond)
	cond.Subscribe(s)
}

func (s *compositeCond) Name() string {
	return s.name
}

func (s *compositeCond) Get() bool {
	for _, cond := range s.conditions {
		if !cond.Get() {
			return false
		}
	}
	return true
}

func (s *compositeCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *compositeCond) OnChange(cond Condition, value bool) {
	if s.value != s.Get() {
		s.set(!s.value)
	}
}

func (s *compositeCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
}

func (s *compositeCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *compositeCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func NewCompositeCondition(name string) *compositeCond {
	cond := &compositeCond{
		observable: NewObservable[Condition, bool](),
		name:       name,
	}

	return cond
}
