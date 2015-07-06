package serverutil

import (
	"mail-test/server"
	"mail-test/client"
	"net"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)
func Serve(s *server.Server,listen net.Listener) (error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,os.Interrupt,os.Kill,syscall.SIGTERM,syscall.SIGHUP)

	for {
		conn,err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer conn.Close()
		
		client := client.NewClient(conn,s)
		go client.Handle()
	}

	_ = <- sig
	return nil
}
func ListenAndServe (s *server.Server,addr string) (error) {
	listen,err := net.Listen("tcp",addr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer listen.Close()
	
	return Serve(s,listen)
}