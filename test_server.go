package main

import (
	"mail-test/server"
	"mail-test/serverutil"
)

func main () {
	s := server.NewServer("localhost:2500")
	serverutil.ListenAndServe(s,"localhost:2500")
}