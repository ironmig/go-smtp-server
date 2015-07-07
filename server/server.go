package server

import (
	"crypto/tls"
)
type AllowAddrFunc func (addr string) (bool)
type Authenticator interface {
	Authenticate(username, password string) (AllowAddrFunc,error)
}
type Server struct {
	Hostname string
	TLSconfig *tls.Config
	Auth Authenticator
}
func NewServer(hostname string) (*Server) {
	return &Server{Hostname:hostname}
}