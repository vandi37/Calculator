package resp

type Created struct {
	Id int `json:"id"`
}

type All struct {
	Expressions []any `json:"expressions"`
}

type Expression struct {
	Expression any `json:"expression"`
}

type Job struct {
	Task any `json:"task"`
}

type ResponseError struct {
	Error string `json:"error"`
}
