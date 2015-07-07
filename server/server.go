package server

import (
	"crypto/tls"
	"io"
	"net/mail"
)
type AllowAddrFunc func (addr string) (bool)
type deliverFunc func (mail.Message) (error)
type Authenticator interface {
	Authenticate(username, password string) (AllowAddrFunc,error)
}
type Server struct {
	Hostname string
	TLSconfig *tls.Config
	Auth Authenticator
	DeliverLocal deliverFunc
	AllowProxy bool //All mail to be sent neither to a local address nor from
}
func WriteMail (m *mail.Message,w io.Writer) (error) {
	var err error
	for key,val := range m.Header { //For all headers, ignores wrapping for now
		_,err = w.Write([]byte(key+": "))
		if err != nil {
			return err
		}
		for i:=0;i<len(val)-1;i++ { //Writes all before last to include semicolon
			w.Write([]byte(val[i]+";"))
		}
		_,err = w.Write([]byte(val[len(val)-1]+"\r\n")) //Writes last one
		if err != nil {
			return err
		}
	}
	_,err = w.Write([]byte("\r\n")) //Skip line to begin body
	if err != nil {
		return err
	}
	_,err = io.Copy(w,m.Body)
	return err
}

func NewServer(hostname string) (*Server) {
	return &Server{Hostname:hostname}
}