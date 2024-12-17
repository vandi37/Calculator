package handler

import "net/http"

func CheckMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			SendJson(w, ResponseError{MethodNotAllowed})
			return
		}
		next(w, r)
	}
}
