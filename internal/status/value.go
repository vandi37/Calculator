package status

import (
	"github.com/vandi37/Calculator/internal/agent/module"
)

type ErrorValue struct{ Error error }

func (v ErrorValue) statusValue() {}

func (v ErrorValue) ForJson() any { return v.Error.Error() }

func (v ErrorValue) GetBack() chan float64 { return nil }

type FinishedValue float64

func (v FinishedValue) statusValue() {}

func (v FinishedValue) ForJson() any { return v }

func (v FinishedValue) GetBack() chan float64 { return nil }

type WaitingValue struct {
	module.Request
	Back chan float64
}

func (v WaitingValue) statusValue() {}

func (v WaitingValue) ForJson() any { return v.Request }

func (v WaitingValue) GetBack() chan float64 { return v.Back }

type ProcessingValue chan float64

func (v ProcessingValue) statusValue() {}

func (v ProcessingValue) ForJson() any { return nil }

func (v ProcessingValue) GetBack() chan float64 { return v }

type NothingValue struct{}

func (v NothingValue) statusValue() {}

func (v NothingValue) ForJson() any { return nil }

func (v NothingValue) GetBack() chan float64 { return nil }
