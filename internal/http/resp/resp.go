package resp

type ResponseOK struct {
	Result float64 `json:"result"`
}

type ResponseError struct {
	Error string `json:"error"`
}
