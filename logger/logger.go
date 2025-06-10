package logger

type EventType int

const (
	_                     = iota
	EventDelete EventType = iota // 1
	EventPut    EventType = iota // 2

)

type Event struct {
	Sequence  uint64    // A unique record ID, in monotonically increasing order
	EventType EventType //A descriptor of action taken; PUT or DELETE
	Key       string    // The key affected by the transaction
	Value     string    // If the event is a EventPut, this will be the value of the transaction
}

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
}
