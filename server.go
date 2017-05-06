package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	incoming chan string
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (client *Client) Read() {
	for {
		line, _ := client.reader.ReadString('\n')
		client.incoming <- line
	}
}

func (client *Client) Write() {
	for data := range client.outgoing {
		client.writer.WriteString(data)
		client.writer.Flush()
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		incoming: make(chan string),
		outgoing: make(chan string),
		reader:   reader,
		writer:   writer,
	}

	client.Listen()

	return client
}

type ChatRoom struct {
	Name     string
	clients  []*Client
	joins    chan net.Conn
	incoming chan string
	outgoing chan string
}

func (chatRoom *ChatRoom) Broadcast(data string) {
	for _, client := range chatRoom.clients {
		client.outgoing <- data
	}
}

func (chatRoom *ChatRoom) Join(connection net.Conn) {
	client := NewClient(connection)
	chatRoom.clients = append(chatRoom.clients, client)
	go func() {
		for {
			chatRoom.incoming <- <-client.incoming
		}
	}()
}

func (chatRoom *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case data := <-chatRoom.incoming:
				chatRoom.Broadcast(data)
			case conn := <-chatRoom.joins:
				chatRoom.Join(conn)
			}
		}
	}()
}
func NewChatRoom() *ChatRoom {
	chatRoom := &ChatRoom{
		Name:     "teste",
		clients:  make([]*Client, 0),
		joins:    make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}
	fmt.Println("Criou")

	chatRoom.Listen()

	return chatRoom
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func BytesToString(data []byte) string {
	return string(data[:len(data)-1])
}

func action(conn net.Conn, roons *[]ChatRoom) {
	// close connection on exit
	defer conn.Close()
	var buf [512]byte
	for {
		// read upto 512 bytes
		_, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		if string(buf[:5]) == "/list" {
			for i := 0; i <= len(*roons); i++ {
				aux := (*roons)[i]
				fmt.Println("Entrou for\n")
				io.WriteString(conn, fmt.Sprint(ChatRoom(aux).Name))
			}

		}
		if string(buf[:7]) == "/create" {
			// read upto 512 bytes
			conn.Read(buf[0:])
			(*roons) = append((*roons), *NewChatRoom())
			println(string(buf[:5]))
		}
		if string(buf[:5]) == "/join" {
			println("deucerto")

		}
	}
}

func main() {
	var (
		//clients  = []Client{}
		ChatRoom = []ChatRoom{}
	)

	listener, _ := net.Listen("tcp", ":4440")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		action(conn, &ChatRoom)
	}
}
