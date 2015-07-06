package main

import (
	"net"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"time"
	"strings"
	"errors"
	"io"
	"net/mail"
	"net/textproto"
	//"bytes"
)
var ErrNoCommand = errors.New("Command not found in messsage")
type Message struct{
	Command string
	Arguments []string
}
type Client struct {
	conn net.Conn
	Writer io.Writer
	DataReader *bufio.Reader
	CommandScanner *bufio.Scanner
}
func (c *Client) Reply (status int,m string) (int,error) {
	msg := fmt.Sprintf("%d %s\r\n", status,m)
	fmt.Println(msg)
	return c.Writer.Write([]byte(msg))
}
func (c *Client) ReplyOk (m string) (int,error) {
	return c.Reply(250,m)
}
func (c *Client) ReadMessage () (Message,error) {
	var m Message
	var err error
	if (c.CommandScanner.Scan()) {
		str := c.CommandScanner.Text()
		m,err = ParseMessage(str)
		if err != nil {
			return m,err
		}
		return m,nil
	} else {
		if err = c.CommandScanner.Err(); err != nil {
			return m, err
		}
		return m,io.EOF
	}
}
func NewClient (c net.Conn) (Client) {
	client := Client{}
	client.conn = c
	client.Writer = client.conn
	client.DataReader = bufio.NewReader(textproto.NewReader(bufio.NewReader(client.conn)).DotReader())
	client.CommandScanner = bufio.NewScanner(client.conn)
	return client
}
func ParseMessage(s string) (Message,error) {
	fmt.Println("Message:",s)
	m := Message{}
	list := strings.Split(s," ")
	if len(list) < 1 {
		return m,ErrNoCommand
	}
	m.Command = list[0]
	m.Arguments = list[1:len(list)]
	return m,nil
}
func main () {
	server()
}
func server () {
	listen,err := net.Listen("tcp","localhost:2500")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen.Close()
	
	close := make(chan os.Signal, 1)
	signal.Notify(close,os.Interrupt,os.Kill,syscall.SIGTERM,syscall.SIGHUP)

	go Listen(listen)

	sig := <- close
	fmt.Println("Clossing",sig)
}
func Listen (listen net.Listener) {
	accept: for {
		conn,err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue accept
		}
		defer conn.Close()
		
		go HandleClient(conn)
	}
}
func HandleClient (conn net.Conn) {
	fmt.Println("Received Connection",conn)
	client := NewClient(conn)
	time.Sleep(2*time.Second)
	_,err := client.Reply(220,"localhost ESMTP ready")
	if (err != nil) {
		fmt.Println(err)
		return
	}
	
	startedMail := false
	var from string
	to := []string{}
	
	msg,scanerr := client.ReadMessage()
	for ; scanerr == nil ; msg,scanerr = client.ReadMessage() {
		switch msg.Command {
			case "EHLO","HELO" :
				_,err = client.ReplyOk("OK")
			case "MAIL":
				if !startedMail {
					startedMail = true
					address,perr := mail.ParseAddress(msg.Arguments[0][5:])
					if perr != nil {
						client.Reply(501,"BAD ADDRESS FORMAT")
						break
					}
					from = address.String()
					_,err = client.ReplyOk("READY TO RECEIVE MAIL FROM " +from)
				} else { //If mail was already started
					client.Reply(503,"MAIL ALREADY STARTED")
				}
			case "RCPT":
				if !startedMail {
					client.Reply(503,"NEED TO ISSUE MAIL COMMAND FIRST")
					break
				}
				
				addr,perr := mail.ParseAddress(msg.Arguments[0][3:])
				if (perr != nil) {
					fmt.Println(perr)
					break
				}
				to = append(to,addr.String())
				_,err = client.ReplyOk("OK READY TO SEND TO "+addr.String())
			case "DATA":
				if !startedMail {
					client.Reply(503,"NEED TO ISSUE MAIL COMMAND FIRST")
					break
				}

				_,err = client.Reply(354,"I'm scared, don't send me data yet")
				if (err != nil) {
					fmt.Println(err)
					break
				}
				
				message, err := mail.ReadMessage(client.DataReader)

				if (err != nil) {
					fmt.Println(err)
					_,err = client.Reply(500,"Error in data transfer")
					break
				}
				
				if _,ok := message.Header["To"]; ok {
					message.Header["To"] = append(message.Header["To"],to...)
				} else {
					message.Header["To"] = to
				}
				message.Header["From"] = []string{from} //Adds previously received from line
				
				//var body bytes.Buffer
				//io.Copy(&body,message.Body)

				_,err = client.ReplyOk("DATA RECEIVED")
			case "RSET":
				startedMail = false;
				from = ""
				to = []string{}
				_,err = client.ReplyOk("RESET")
			case "QUIT":
				_,err = client.Reply(221,"CLOSSING CONNECTION")
				return
			default:
				_,err = client.Reply(502,"UNRECOGNIZED COMMAND")
				return
		}
	}
	fmt.Println("Stopped scanning",scanerr)
	return
}