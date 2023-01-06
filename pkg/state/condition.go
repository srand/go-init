package state

type Condition interface {
	Observable[Condition, bool]
	Name() string
	Get() bool
}
