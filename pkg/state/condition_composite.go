package state

type compositeCond struct {
	Condition
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
	or         bool
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
	return s.value
}

func (s *compositeCond) get() bool {
	if s.or {
		for _, cond := range s.conditions {
			if cond.Get() {
				return true
			}
		}
		return false
	} else {
		for _, cond := range s.conditions {
			if !cond.Get() {
				return false
			}
		}
		return true
	}
}

func (s *compositeCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *compositeCond) OnChange(cond Condition, value bool) {
	s.set(s.get())
}

func (s *compositeCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
	observer.OnChange(s, s.Get())
}

func (s *compositeCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *compositeCond) Publish() {
	s.observable.Publish(s, s.value)
}

func NewCompositeCondition(name string, or bool, def bool) *compositeCond {
	cond := &compositeCond{
		observable: NewObservable[Condition, bool](),
		name:       name,
		or:         or,
		value:      def,
	}

	cond.Subscribe(&ConditionLogger{})

	return cond
}
