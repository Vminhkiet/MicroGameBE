package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9090")

	if err != nil {
		fmt.Printf("Loi ket noi %v\n", err)
		return
	}

	defer conn.Close()

	go receiveMessage(conn)

	chatServer(conn)
}

func receiveMessage(conn net.Conn) {
	buffer := bufio.NewReader(conn)

	for {
		message, err := buffer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi doc buffer %v\n", err)
			return
		}

		fmt.Print("📨 Server gửi: ", message)
	}
}

func chatServer(conn net.Conn) {
	buffer := bufio.NewReader(os.Stdin)

	for {
		message, err := buffer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi doc buffer %v\n", err)
			return
		}

		conn.Write([]byte(message))
	}
}
