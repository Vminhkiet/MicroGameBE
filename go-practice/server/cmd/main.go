package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	clients = make(map[string]net.Conn)
	mu      sync.RWMutex
)

func main() {

	listener, err := net.Listen("tcp", ":9090")

	if err != nil {
		fmt.Printf("Loi %v\n", err)
		return
	}

	defer listener.Close()

	go readStdin()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()

	mu.Lock()
	clients[addr] = conn
	mu.Unlock()

	buffer := bufio.NewReader(conn)

	for {
		message, err := buffer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			break
		}
		sendMessage(conn, message)
	}

	mu.Lock()
	delete(clients, addr)
	mu.Unlock()
}

func readStdin() {
	writer := bufio.NewReader(os.Stdin)

	for {
		message, err := writer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi %v\n", err)
			return
		}
		sendMessage(nil, message)
	}
}

func sendMessage(con net.Conn, message string) {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Printf("Server nhan thong tin: %v\n", message)

	for addr, conn := range clients {
		if con == conn {
			continue
		}

		_, err := conn.Write([]byte("Server: " + message))
		if err != nil {
			fmt.Printf("Không gửi được tới %s: %v\n", addr, err)
		}
		fmt.Printf("Server gui den %s: %v\n", addr, err)
	}
}
