package status

import (
	"fmt"

	"github.com/vandi37/Calculator/internal/agent/module"
)

type ErrorValue struct{ Error error }

func (v ErrorValue) statusValue() {}

func (v ErrorValue) ForJson() any { return v.Error.Error() }

func (v ErrorValue) GetBack() chan float64 { return nil }

func (v ErrorValue) String() string { return v.Error.Error() }

func (v ErrorValue) GerRequest() *module.Request { return nil }

type FinishedValue float64

func (v FinishedValue) statusValue() {}

func (v FinishedValue) ForJson() any { return v }

func (v FinishedValue) GetBack() chan float64 { return nil }

func (v FinishedValue) String() string { return fmt.Sprint(float64(v)) }

func (v FinishedValue) GerRequest() *module.Request { return nil }

type WaitingValue struct {
	module.Request
	Back chan float64
}

func (v WaitingValue) statusValue() {}

func (v WaitingValue) ForJson() any { return v.Request }

func (v WaitingValue) GetBack() chan float64 { return v.Back }

func (v WaitingValue) String() string { return v.Request.String() }

func (v WaitingValue) GerRequest() *module.Request { return &v.Request }

type ProcessingValue struct {
	module.Request
	Back chan float64
}

func (v ProcessingValue) statusValue() {}

func (v ProcessingValue) ForJson() any { return v.Request }

func (v ProcessingValue) GetBack() chan float64 { return v.Back }

func (v ProcessingValue) String() string { return v.Request.String() }

func (v ProcessingValue) GerRequest() *module.Request { return &v.Request }

type NothingValue struct{}

func (v NothingValue) statusValue() {}

func (v NothingValue) ForJson() any { return nil }

func (v NothingValue) GetBack() chan float64 { return nil }

func (v NothingValue) String() string { return "void" }

func (v NothingValue) GerRequest() *module.Request { return nil }
