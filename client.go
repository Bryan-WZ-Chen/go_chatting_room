package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP    string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// Create client instance
	client := &Client {
		ServerIP: serverIp,
		ServerPort: serverPort,
		flag: 999, // some non-zero value here
	}

	// connect to server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIP, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn
	// return instance
	return client
}

func (client *Client) ShowOnlineUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err: ", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.ShowOnlineUsers()
	fmt.Println("Select a user that you want to chat with, type exit to leave")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("please enter some content ..., type exit to leave")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err: ", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("please enter some content ..., type exit to leave")
			fmt.Scanln(&chatMsg)
		}

		client.ShowOnlineUsers()
		fmt.Println("Select a user that you want to chat with, type exit to leave")
		fmt.Scanln(&remoteName)
	}

}

func (client *Client) PublicChat() {
	var chatMsg string
	
	// prompt user to enter message
	fmt.Println("please enter some content ..., type exit to leave")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// send msg to server
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err: ", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("please enter some content ..., type exit to leave")
		fmt.Scanln(&chatMsg)
	}

}

// goroutine that reads message from server and display on the client side
func(client *Client) DealResponse() {
	// blocking listening
	io.Copy(os.Stdout, client.conn)

	// equivalent
	/*
	for {
		buf := make([]byte, 4096)
		clent.conn.Read(buf)
		fmt.Println(buf)
	}
	*/
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1. chatting in public mode")
	fmt.Println("2. chatting with someone (private mode)")
	fmt.Println("3. update user name")
	fmt.Println("0. exit")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("Please enter valid number between 0 ~ 3")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("Enter new user name")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {}

		// do different tasks based on flag
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Configure server ip (default address is 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "Configure server port (default port is 8888)")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>> failed to connect to server ...")
		return
	}

	fmt.Println(">>>> connection successful ...")

	go client.DealResponse()

	// Do client side business logic here
	client.Run()
}