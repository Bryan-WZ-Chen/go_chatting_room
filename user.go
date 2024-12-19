package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// Create a user instance
func NewUser(conn net.Conn, server *Server) *User {
	// We use address as a default name for the user
	userAddr := conn.RemoteAddr().String()


	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
		server: server,
	}

	// start a goroutine that monitors user's channel
	go user.ListenMessage()
	return user
}

func (this *User) Online() {
	// A new user is currently online so we update the map
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// broadcasting that the user is currently online
	this.server.BroadCast(this, "is currently online")	
}

func (this *User) Offline() {
	// A user is offline so we remove it from the map
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// broadcasting that the user is currently online
	this.server.BroadCast(this, "is offline")
}

func (this *User)SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	// Assume "who" is the command that is used to list all online users
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "is online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		// determine whether newName exists in the map
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("Current user name already being used\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("Update user name to: " + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// message format: to|${name}|${message}

		// 1. retrieve username
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("Invalid message format, should be to|${name}|${message}\n")
			return
		}
		// 2. retrieve User struct according to the username
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("targeted username doesn't exist\n")
			return
		}

		// 3. retrieve message and send message to the targeted User
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("no content, please resend message!\n")
			return
		}
		remoteUser.SendMsg("From " + this.Name + ": " + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}

// Monitor whether user's channel has data and send to user if there's data in the channel
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
