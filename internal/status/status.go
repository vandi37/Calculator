package status

import (
	"fmt"

	"github.com/vandi37/Calculator/internal/agent/module"
)

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
		return "Nothing"
	case Processing:
		return "Processing"
	case Waiting:
		return "Waiting"
	case Error:
		return "Error"
	case Finished:
		return "Finished"
	default:
		return "Unknown status"
	}
}

type StatusValue interface {
	ForJson() any
	GerRequest() *module.Request
	GetBack() chan float64
	statusValue()
	fmt.Stringer
}

type Status struct {
	Level      StatusLevel
	Value      StatusValue
	Expression string
}

type statusJson struct {
	Id         int    `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"`
	Value      any    `json:"value"`
}

func (s Status) ForJson(id int) any {
	return statusJson{id, s.Expression, s.Level.String(), s.Value.ForJson()}
}

func New(status StatusLevel, value StatusValue, expression string) Status {
	return Status{status, value, expression}
}
