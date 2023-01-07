package state

type refCond struct {
	registry   ReferenceRegistry
	observable *SimpleObservable[Condition, bool]
	name       string
	value      bool
}

func NewConditionRef(registry ReferenceRegistry, name string) Condition {
	cond := &refCond{
		registry:   registry,
		observable: NewObservable[Condition, bool](),
		name:       name,
		value:      false,
	}
	registry.SubscribeReference(name, cond)
	return cond
}

func (s *refCond) Name() string {
	return s.name
}

func (s *refCond) Get() bool {
	return s.value
}

func (s *refCond) set(value bool) {
	if s.value != value {
		s.value = value
		s.Publish()
	}
}

func (s *refCond) Subscribe(observer Observer[Condition, bool]) {
	s.observable.Subscribe(observer)
	observer.OnChange(s, s.Get())
}

func (s *refCond) Unsubscribe(observer Observer[Condition, bool]) {
	s.observable.Unsubscribe(observer)
}

func (s *refCond) Publish() {
	s.observable.Publish(s, s.Get())
}

func (s *refCond) OnChange(cond Condition, value bool) {
	s.set(value)
}

func (s *refCond) OnReferenceFound(name string, obj any) {
	if s.name != name {
		return
	}

	if cond, ok := obj.(Condition); ok {
		cond.Subscribe(s)
		s.set(cond.Get())
	}
}

func (s *refCond) OnReferenceLost(name string, obj any) {
	if s.name != name {
		return
	}

	if cond, ok := obj.(Condition); ok {
		cond.Unsubscribe(s)
		s.set(false)
	}
}
