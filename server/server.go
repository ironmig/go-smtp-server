package server

import ()

type Server struct {
	Hostname string
}
func NewServer(hostname string) (*Server) {
	return &Server{Hostname:hostname}
}