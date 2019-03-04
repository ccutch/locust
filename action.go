package locust

type actionType int

type action struct {
	RequestID string
	Type      actionType
	Target    string
}

const (
	kill actionType = iota
	ping
)
