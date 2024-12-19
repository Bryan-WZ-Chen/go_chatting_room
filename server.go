package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// map that records users that are currently online
	OnlineMap map[string]*User
	// map is a global object so we may need to add a lock for it
	mapLock sync.RWMutex

	// channel used to broadcast
	Message chan string
}

// Create a server instance
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:   ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
}

func (this *Server)ListenMsg() {
	for {
		msg := <- this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server)Handler(conn net.Conn) {

	user := NewUser(conn, this)

	user.Online()
	isAlive := make(chan bool)
	// receive message from user (client) and broadcast
	go func() {
		buf := make([]byte, 4096)
		
		for {
			n, err := conn.Read(buf)
			
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}

			// remove last character "\n"
			msg := string(buf[: n - 1])
			user.DoMessage(msg)
			isAlive <-true
		}
	}()

	// blocking current goroutine
	for {
		select{
			case <-time.After(time.Second * 300):
				user.SendMsg("timeout!! Reconnection needed!")
				user.Offline()
				// close a channel
				close(user.C)
				user.conn.Close() //conn.Close()
				return
			case <- isAlive:
		}
	}
}

// Activate a server
func (this *Server) Start() {
	// socket listening
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	} else {
		fmt.Println("server starts running ...")
	}

	// close listening
	defer listener.Close()

	// start a goroutine that monitors server's channel and broadcasts to all users
	go this.ListenMsg()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
			continue
		}
		// handler
		go this.Handler(conn)
	}
}