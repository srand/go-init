package state

type Trigger interface {
	Name() string
	Eval()
}

type actionTrigger struct {
	name  string
	cond  Condition
	up    Action
	down  Action
	value bool
}

func (t *actionTrigger) Name() string {
	return t.name
}

func (t *actionTrigger) Eval() {
	t.OnChange(t.cond, t.cond.Get())
}

func (t *actionTrigger) OnChange(cond Condition, asserted bool) {
	// log.Println("?", t.name, asserted)

	// Only trigger on transitions
	if t.value != asserted {
		t.value = asserted

		if asserted {
			if t.up != nil {
				t.up.Run()
			}
		} else {
			if t.down != nil {
				t.down.Run()
			}
		}
	}
}

func NewActionTrigger(name string, cond Condition, up, down Action) Trigger {
	trigger := &actionTrigger{
		name:  name,
		cond:  cond,
		up:    up,
		down:  down,
		value: false,
	}

	cond.Subscribe(trigger)

	return trigger
}
