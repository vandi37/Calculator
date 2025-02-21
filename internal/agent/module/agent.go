package module

import "github.com/vandi37/Calculator/pkg/parsing/tree"

type Request struct {
	Id              int          `json:"id"`
	Arg1            float64      `json:"arg1"`
	Arg2            float64      `json:"arg2"`
	Operation       tree.ExprSep `json:"operation"`
	OperationTimeMs int          `json:"operation_time"`
}

type Post struct {
	Id     int     `json:"id"`
	Result float64 `json:"arg1"`
}
