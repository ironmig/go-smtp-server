package main

import (
	"fmt"
	"net/smtp"
	"net"
	"os"
	"io"
)
func main() {
	
	conn,err := net.Dial("tcp","localhost:2500")
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := smtp.NewClient(conn,"localhost:2500")
	if err != nil {
		fmt.Println(err)
		return
	}
   defer client.Close()
	
	err = client.Hello("localhost")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Heloed")
	
	err = client.Mail("test@localhost")
	if err != nil {
		fmt.Println(err)
		return
	}	
	fmt.Println("Mail suc")
	
	err = client.Rcpt("afair@localhost")
	if err != nil {
		fmt.Println(err)
		return
	}	
	fmt.Println("Recept succ")
	
	err = client.Rcpt("negro@gmail.com")
	if err != nil {
		fmt.Println(err)
		return
	}	
	fmt.Println("Recept succ")
	
	write,err := client.Data()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Data ready")
	
	file,err := os.Open("test.eml")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	length,err := io.Copy(write,file)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Write success",length)
	
	err = write.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Data Sent")
	
	err = client.Reset()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Reset succ")
	
	
	err = client.Quit()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Quit succ")	
	
	/*
	if len(os.Args) > 1 {
		file,err := os.Open(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		
		conn,err := net.Dial("tcp4","localhost:2500")
		if err != nil {
			fmt.Println(err)
			return
		}
		
		SendMail(conn,file)
	} else {
		fmt.Println("No file")
	}
	*/
}
/*
func SendMail(conn net.Conn,message io.Reader) {
	fmt.Println("Received Conn",conn.RemoteAddr())
	client,err :=  smtp.NewClient(conn,"localhost:2500")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("conn est")
	err = client.Hello("test")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.Rcpt("bob")
	if err != nil {
		fmt.Println(err)
		return
	}
}
*/