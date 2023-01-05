package services

type Condition interface {
	IsMet() bool
}
