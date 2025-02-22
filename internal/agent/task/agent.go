package task

import (
	"github.com/vandi37/Calculator/internal/agent/do"
	"github.com/vandi37/Calculator/internal/agent/module"
)

func Task(req module.Request, sendBack func(int, float64, module.Request)) {
	sendBack(req.Id, do.Do(req), req)
}
