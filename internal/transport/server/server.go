package server

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	http.Server
}

func New(handler http.Handler, port int) *Server {
	return &Server{http.Server{Addr: fmt.Sprint(":", port), Handler: handler}}
}

func (s *Server) Run(logger *zap.Logger) error {
	logger.Info("server running", zap.String("addr", s.Addr))
	err := s.ListenAndServe()
	return err
}
