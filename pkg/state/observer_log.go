package state

import "log"

type StateLogger struct{}

func (o *StateLogger) OnChange(obj State[string], val string) {
	log.Println(obj.Name(), val)
}

type ConditionLogger struct{}

func (o *ConditionLogger) OnChange(obj Condition, val bool) {
	log.Println(obj.Name(), val)
}
