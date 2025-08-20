package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Dial("tcp", "localhost:9090")

	if err != nil {
		fmt.Printf("Loi %v\n", err)
		return
	}

	defer listener.Close()

	go receiveMessage(listener)

	buffer := bufio.NewReader(os.Stdin)

	for {
		text, err := buffer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			return
		}

		_, err = listener.Write([]byte(text))

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			return
		}

		fmt.Printf("Gui den server %v\n", text)
	}
}

func receiveMessage(conn net.Conn) {

	buffer1 := bufio.NewReader(conn)

	for {
		message, err := buffer1.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			return
		}

		fmt.Printf("Nhan tu server %v\n", message)
	}
}
