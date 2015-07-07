package main

import (
	"mail-test/server"
	"mail-test/serverutil"
	"fmt"
	"crypto/tls"
	"net"
	"net/mail"
	"os"
)
const password = "password"
type InsecureAuth struct{}
func (i InsecureAuth) Authenticate (u, p string) (server.AllowAddrFunc,error) {
	return func (a string) (bool) {
		fmt.Println(a,u)
		return a == u && password == p
	},nil
}
func DeliverToFileExample(m mail.Message) (error) {
	file,err := os.OpenFile("testing.eml",os.O_RDWR,os.ModeType)
	if err !=  nil {
		fmt.Println(err)
		return err
	}
	err = server.WriteMail(&m,file)
	if err !=  nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func main () {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	config := tls.Config{Certificates:[]tls.Certificate{cert},InsecureSkipVerify:true}
	
	listen,err := net.Listen("tcp","localhost:2500")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	s := server.NewServer("localhost")
	s.TLSconfig = &config
	s.Auth = InsecureAuth{}
	s.DeliverLocal = DeliverToFileExample
	serverutil.Serve(s,listen)
}