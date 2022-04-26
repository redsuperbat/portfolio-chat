package events

type Event interface {
	Valid() error
}
