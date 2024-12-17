package server

import (
	"fmt"
	"net/http"
)

type Server struct {
	http.Server
}

func New(handler http.Handler, port int) *Server {
	return &Server{http.Server{Addr: fmt.Sprint(":", port), Handler: handler}}
}

func (s *Server) Run() error {
	err := s.ListenAndServe()
	return err
}
