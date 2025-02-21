package status

type StatusLevel int

const (
	Processing StatusLevel = (10 - iota) >> iota // why this start idk
	Waiting
	Error
	Finished
	Nothing = -1
)

func (l StatusLevel) String() string {
	switch l {
	case Nothing:
		return "[nothing]"
	case Processing:
		return "[processing]"
	case Waiting:
		return "[waiting]"
	case Error:
		return "[error]"
	case Finished:
		return "[finished]"
	default:
		return "[unknown status]"
	}
}

type StatusValue interface {
	ForJson() any
	GetBack() chan float64
	statusValue()
}

type Status struct {
	Level StatusLevel
	Value StatusValue
}

type statusJson struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
	Value  any    `json:"value"`
}

func (s Status) ForJson(id int) any {
	return statusJson{id, s.Level.String(), s.Value.ForJson()}
}

func New(status StatusLevel, value StatusValue) Status {
	return Status{status, value}
}
