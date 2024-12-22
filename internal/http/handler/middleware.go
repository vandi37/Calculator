package handler

import (
	"net/http"

	"github.com/vandi37/Calculator/internal/http/resp"
)

func CheckMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			SendJson(w, resp.ResponseError{Error: MethodNotAllowed})
			return
		}
		next(w, r)
	}
}

func CheckPath(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			NotFoundHandler(w, r)
			return
		}
		next(w, r)
	}
}

func ContentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next(w, r)
	}
}
