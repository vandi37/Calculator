package server

import (
	"fmt"
	"net/http"
)

// The server
type Server struct {
	http.Server
}

// Creates a new server
func New(handler http.Handler, port int) *Server {
	return &Server{http.Server{Addr: fmt.Sprint(":", port), Handler: handler}}
}

// Runs server
func (s *Server) Run() error {
	err := s.ListenAndServe()
	return err
}
