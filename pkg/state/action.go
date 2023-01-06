package state

type Action interface {
	Name() string
	Run() error
}

type ActionFunc func() error

type simpleAction struct {
	name string
	run  ActionFunc
}

func (a *simpleAction) Name() string {
	return a.name
}

func (a *simpleAction) Run() error {
	return a.run()
}

func NewAction(name string, run ActionFunc) Action {
	return &simpleAction{name: name, run: run}
}
