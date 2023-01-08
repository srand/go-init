package state

type Condition interface {
	Observable[Condition, bool]
	Name() string
	Get() bool
}

type ReferenceObserver interface {
	OnReferenceFound(string, any)
	OnReferenceLost(string, any)
}

type ReferenceRegistry interface {
	FindReference(string) any
	SubscribeReference(string, ReferenceObserver)
	UnsubscribeReference(string, ReferenceObserver)
}
