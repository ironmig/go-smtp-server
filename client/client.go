package client

import (
	"net"
	"fmt"
	"bufio"
	"time"
	"strings"
	"net/mail"
	"net/textproto"
	"mail-test/server"
	"io"
	"errors"
	"crypto/tls"
	"encoding/base64"
)
var ErrNoCommand = errors.New("No command found in message")
type Message struct {
	Command string
	Arguments []string
}
type Address struct {
	Name string
	User string
	Hostname string
}
func NewAddress(a mail.Address) (Address) {
	list := strings.Split(a.Address,"@")
	return Address{Name:a.Name,User:list[0],Hostname:list[1]}
}
func (a Address) String () (string) {
	str := ""
	if a.Name != "" {
		str = str + a.Name + " "
	}
	str = str + "<"+a.User+"@"+a.Hostname+">"
	return str
}
func (a Address) Address() (string) {
	return a.User+"@"+a.Hostname
}
type Addresses []Address
func (a Addresses) Strings () ([]string) {
	res := []string{}
	for _,addr := range a {
		res = append(res,addr.String())
	}
	return res
}
func (a Addresses) Addresses() ([]string) {
	res := []string{}
	for _,addr := range a {
		res = append(res,addr.Address())
	}
	return res
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
type Client struct {
	conn net.Conn
	server *server.Server
	Writer io.Writer
	DataReader *bufio.Reader
	CommandScanner *bufio.Scanner
	Helloed bool
	AllowAddress server.AllowAddrFunc
	startedMail bool
	isFromLocal bool
	isToLocal bool
	to Addresses
	from Address
	message mail.Message
}
func NewClient (c net.Conn,s *server.Server) (Client) {
	client := Client{}
	client.server = s
	client.conn = c
	client.initInterfaces()
	return client
}
func (c *Client) initInterfaces () () {
	c.Writer = c.conn
	c.DataReader = bufio.NewReader(textproto.NewReader(bufio.NewReader(c.conn)).DotReader())
	c.CommandScanner = bufio.NewScanner(c.conn)
}
func (c *Client) Reply (status int,m string) (int,error) {
	msg := fmt.Sprintf("%d %s\r\n", status,m)
	fmt.Println(msg)
	return c.Writer.Write([]byte(msg))
}
func (c *Client) ReplyCode (status int ) (int,error) {
	msg := fmt.Sprintf("%d \r\n", status)
	fmt.Println(msg)
	return c.Writer.Write([]byte(msg))
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
func (client *Client) Reset() {
	client.startedMail = false;
	client.from = Address{}
	client.to = []Address{}
	client.message = mail.Message{}
}
func (client *Client) FlushReader() {
	bufio.NewReader(client.conn).Reset(bufio.NewReader(client.conn))
}
func (client *Client) HandleHelo (msg Message) () {
	send := fmt.Sprintf("%d-%s\r\n%d-%s\r\n%d %s\r\n",250,client.server.Hostname,250,"AUTH PLAIN",250,"STARTTLS")
	fmt.Println(send)
	_,err := client.Writer.Write([]byte(send))
	if err != nil {
		fmt.Println(err)
		return
	}
	client.Helloed = true;
}
func (client *Client) HandleMail (msg Message) () {
	if !client.startedMail {
		addr,perr := mail.ParseAddress(msg.Arguments[0][5:])
		if perr != nil {
			client.Reply(501,"BAD ADDRESS FORMAT")
			return
		}
		client.from = NewAddress(*addr)
		if (client.server.Hostname == client.from.Hostname) { //If message is from server's host, require authentiction now
			if client.server.Auth != nil { //If server has authentication
				if client.AllowAddress == nil { //If the client has not Authed
					client.Reply(503,"Not authenticated")
					return
				}
				if !client.AllowAddress(client.from.Address()) { //Check if from address is valid for credentials
					client.Reply(503,"not from you")
					return
				}
			}
			client.isFromLocal = true
		}

		client.startedMail = true

		client.ReplyCode(250)
	} else {
		client.Reply(503,"MAIL ALREADY STARTED")
	}
}
func (client *Client) HandleRecipient (msg Message) () {
	if !client.startedMail {
		client.Reply(503,"NEED TO ISSUE MAIL COMMAND FIRST")
		return
	}
	maddr,perr := mail.ParseAddress(msg.Arguments[0][3:])
	if (perr != nil) {
		fmt.Println(perr)
		client.Reply(503,"Error parsing address")
		return
	}
	addr := NewAddress(*maddr)
	if (addr.Hostname == client.server.Hostname) {
		client.isToLocal = true
	}
	client.to = append(client.to,addr)
	client.ReplyCode(250)
}
func (client *Client) HandleData (msg Message) {
	if !client.startedMail {
		client.Reply(503,"NEED TO ISSUE MAIL COMMAND FIRST")
		return
	}
	
	if !client.isFromLocal && !client.isToLocal && !client.server.AllowProxy {
		client.Reply(503,"PROXY MAIL NOT ALLOWED")
		return
	}
	
	_,err := client.ReplyCode(354)
	if err != nil {
		fmt.Println(err)
		return
	}
	message, err := mail.ReadMessage(client.DataReader)
	
	defer client.FlushReader()
	
	if err != nil {
		fmt.Println(err)
		client.Reply(500,"Error in data transfer")
		return
	}

	message.Header["To"] = client.to.Strings()
	message.Header["From"] = []string{client.from.String()}
	
	if (client.isFromLocal) {
		err = client.server.DeliverLocal(*message)
		if err != nil {
			client.Reply(500,"Error transfer email locally")
			return
		}
	} else {
		
	}

	client.ReplyCode(250)
}
func (client *Client) HandleReset (msg Message) () {
	client.Reset()
	client.ReplyCode(250)
}
func (client *Client) HandleStartTLS (msg Message) () {
	if client.server.TLSconfig != nil {
		client.conn = tls.Server(client.conn,client.server.TLSconfig)
		client.ReplyCode(220)
		client.initInterfaces() //Reset readers and writers to be over tls
		return
	}
	client.Reply(502,"NOT SUPPORTED")
}
func (client *Client) HandleAuth (msg Message) () {
	if client.server.Auth == nil { //If auth is not supported
		client.Reply(502,"NO AUTH")
		return
	}
	switch msg.Arguments[0] {
		case "PLAIN":
			decoded,err := base64.StdEncoding.DecodeString(msg.Arguments[1])
			if err != nil {
				fmt.Println(err)
				client.Reply(502,err.Error())
				return
			}	
			info := strings.Split(string(decoded),"\x00")
			allowaddr,err := client.server.Auth.Authenticate(info[1],info[2])
			if err != nil {
				fmt.Println(err)
				client.Reply(502,err.Error())
				return
			}
						
			client.AllowAddress = allowaddr
			client.ReplyCode(235) //If successfull authenticated
		default:
			client.Reply(502,"UNSUPORTED AUTH METHOD")
	}
}
func (client *Client) Handle () {
	fmt.Println("Received client")
	time.Sleep(2*time.Second)
	_,err := client.Reply(220,"localhost ESMTP ready")
	if (err != nil) {
		fmt.Println(err)
		return
	}

	msg,scanerr := client.ReadMessage()
	for ; scanerr == nil ; msg,scanerr = client.ReadMessage() {
		switch msg.Command {
			case "EHLO","HELO" : client.HandleHelo(msg)
			case "MAIL": client.HandleMail(msg)
			case "RCPT": client.HandleRecipient(msg)
			case "DATA": client.HandleData(msg)
			case "RSET": client.HandleReset(msg)
			case "QUIT": 
				_,err = client.Reply(221,"CLOSSING CONNECTION")
				return //Leave handle func, allow connection to be clossed and garb collected
			case "STARTTLS": client.HandleStartTLS(msg)
			case "AUTH": client.HandleAuth(msg)
			default:
				_,err = client.Reply(502,"UNRECOGNIZED COMMAND")
		}
	}
	fmt.Println("Stopped scanning",scanerr)
	return
}