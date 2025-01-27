package module

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

var (
	conn                 net.Conn
	HistoryMessage       []byte
	HistoryMessageLength int
	Key                  []byte
)

func init() {
	HistoryMessage = make([]byte, 1024)
	Key = make([]byte, 32)
	HistoryMessageLength = 0
}

func ClientStart() {
	tcpAddress, err := net.ResolveTCPAddr("tcp", ServerPort)
	if err != nil {
		fmt.Println("can not resolve tcp address", err)
	}
	fmt.Println("tcp address", tcpAddress)
	conn, err = net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		fmt.Println("连接服务器失败:", err)
		return
	}
	fmt.Println("Connected to server:", conn.RemoteAddr())
	fmt.Println("输入 '/help' 查看帮助信息！")
	// get the key 获得从服务端传来的密钥
	conn.Read(Key)
	fmt.Println("建立了与服务器的安全连接 !!!")
	go ListenBroadcast()
	go ListenUserInput()
	pause()
}
func ListenBroadcast() {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Server closed")
			os.Exit(0)
		}
		if n != HistoryMessageLength && string(buf[:n]) != string(HistoryMessage[:HistoryMessageLength]) {
			// TODO : add decry data
			bs, err := decrypt(buf[:n], Key)
			Must(err)
			fmt.Print(string(bs))
			fmt.Print(">")
		}
	}
}

func SendMessages() {
	// conn.Write(HistoryMessage[:HistoryMessageLength])
	bs, err := encrypt(Chat(HistoryMessage[:HistoryMessageLength]).ToBuffer(), Key)
	if err != nil {
		fmt.Println("Error encrypting message:", err)
		return
	}
	conn.Write(bs)
}

func ListenUserInput() {
	for {
		fmt.Print(">")
		HistoryMessageLength, _ = os.Stdin.Read(HistoryMessage)
		if HistoryMessage[0] == '/' {
			ParseCommand()
			continue
		}
		SendMessages()
	}
}

// TODO
func ParseCommand() {
	//判断操作系统是
	var command string
	if runtime.GOOS == "windows" {
		command = string(HistoryMessage[1 : HistoryMessageLength-2])
	} else {
		command = string(HistoryMessage[1 : HistoryMessageLength-1])
	}
	words := strings.Split(command, " ")
	switch words[0] {
	case "quit":
		os.Exit(0)
	case "help":
		fmt.Println("/help : get more info about command")
		fmt.Println("/quit : quit the chatroom")
		fmt.Println("/list : list all the users")
		fmt.Println("/rename : rename yourself")
		// fmt.Println("/private : send private message")
	case "list":
		list := List()
		bs, err := encrypt(list.ToBuffer(), Key)
		Must(err)
		conn.Write(bs)
	case "rename":
		if len(words) == 1 {
			fmt.Println("please input your new name ")
			fmt.Println("/rename KM911")
			return
		}

		conn.Write(MustEncrypt(Rename([]byte(words[1])).ToBuffer(), Key))
	// case "private":
	// 	fmt.Println("send private message")
	default:
		fmt.Println("Invalid command")
	}
}
