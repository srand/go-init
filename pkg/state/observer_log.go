package state

import "log"

type StateLogger[T string | int] struct{}

func (o *StateLogger[T]) OnChange(obj State[T], val T) {
	log.Println("=", obj.Name(), val)
}

type ConditionLogger struct{}

func (o *ConditionLogger) OnChange(obj Condition, val bool) {
	if obj.Name() != "" {
		// log.Println(obj.Name(), val)
	}
}
