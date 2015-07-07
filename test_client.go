package main

import (
	"fmt"
	"net/smtp"
	//"net"
	"os"
	"io"
	"crypto/tls"
	"net"
)
func main() {
	config := tls.Config{InsecureSkipVerify:true}
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
	
	err = client.StartTLS(&config)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("TlS success")	
	
	auth := smtp.PlainAuth("","kma1660@localhost","password","localhost:2500")
	err = client.Auth(auth)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Auth success")

	err = client.Mail("kma1660@localhost")
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
	
}