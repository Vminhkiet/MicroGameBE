package main

import (
	"ChatServer/enum"
	"ChatServer/model"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	clients = make(map[string]*model.Client)
	mul     sync.RWMutex
)

func main() {
	listener, err := net.Listen("tcp", ":9090")

	if err != nil {
		fmt.Printf("Loi mo cong %v\n", err)
		return
	}

	defer listener.Close()

	go readStdin()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Loi ket noi %v\n", err)
			continue
		}

		go mainHandle(conn)
	}

}

func readStdin() {
	writer := bufio.NewReader(os.Stdin)

	for {
		message, err := writer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi doc du lieu %v\n", err)
			return
		}

		sendMessage(nil, nil, message)
	}
}

func mainHandle(conn net.Conn) {

	addr := conn.RemoteAddr().String()
	mul.Lock()
	clients[addr] = &model.Client{
		Conn:     conn,
		Username: "",
		Role:     enum.User,
		Muted:    false,
		Kicked:   false,
	}
	mul.Unlock()

	defer func() {
		conn.Close()
		mul.Lock()
		delete(clients, addr)
		mul.Unlock()
		fmt.Printf("👋 Đã ngắt kết nối: %v\n", addr)
	}()

	buffer := bufio.NewReader(conn)

	for {
		message, err := buffer.ReadString('\n')

		if err != nil {
			fmt.Printf("Loi doc du lieu %v\n", err)
			break
		}

		fmt.Printf("Nhan message: %v\n", message)

		go processMessage(conn, message)
	}
}

func processMessage(conn net.Conn, message string) {

	if getKick(conn) {
		return
	}

	role := getRole(conn)

	loginUser(conn, message)

	if role == "user" {
		functionalUser(conn, message)
	}

	if role == "admin" {
		functionalUser(conn, message)
		functionalAdmin(conn, message)
	}

}

func getRole(conn net.Conn) enum.Role {
	if conn == nil {
		return enum.User
	}

	for _, client := range clients {
		if conn == client.Conn {
			return client.Role
		}
	}

	return enum.User
}

func getKick(conn net.Conn) bool {
	if conn == nil {
		return false
	}

	for _, client := range clients {
		if conn == client.Conn {
			return client.Kicked
		}
	}

	return false
}

func getMute(conn net.Conn) bool {
	if conn == nil {
		return false
	}

	for _, client := range clients {
		if conn == client.Conn {
			return client.Muted
		}
	}

	return false
}

func loginUser(conn net.Conn, message string) {
	if strings.HasPrefix(message, "/login") {
		parts := strings.SplitN(message, " ", 3)

		if len(parts) == 3 {
			username := parts[1]
			role := parts[2]

			if role != "user" && role != "admin" {
				return
			}

			mClient := findClientByConn(conn)

			if mClient != nil {
				mClient.Username = username
				mClient.Role = enum.Role(role)

				fmt.Printf("Cap nhat thanh cong %v\n", mClient.Username)
			}
		}
	}
}

func functionalUser(conn net.Conn, message string) {
	if strings.HasPrefix(message, "/msg") {
		parts := strings.SplitN(message, " ", 3)

		if len(parts) == 3 {
			if getMute(conn) {
				return
			}
			username := parts[1]
			content := parts[2]

			clientQuery := findClientByUsername(username)

			if clientQuery != nil {
				sendMessage(nil, clientQuery.Conn, clientQuery.Username+" gửi: "+content)
			} else {
				fmt.Printf("Username khong ton tai %v\n", username)
			}
		}
	} else if strings.HasPrefix(message, "/list") {
		if strings.ReplaceAll(message, " ", "") == "/list" {
			msg := getClientOnline()

			sendMessage(nil, conn, msg)
		}
	} else if strings.HasPrefix(message, "/quit") {
		if strings.TrimSpace(message) == "/quit" {
			conn.Write([]byte("Bạn đã thoát khỏi server.\n"))

			conn.Close()
			return
		}
	}
}

func functionalAdmin(conn net.Conn, message string) {
	if strings.HasPrefix(message, "/broadcast") {
		parts := strings.SplitN(message, " ", 2)

		if len(parts) == 2 {
			content := parts[1]

			sendMessage(conn, nil, content)
		}
	} else if strings.HasPrefix(message, "/mute") {
		parts := strings.Fields(message)

		if len(parts) == 2 && parts[0] == "/muted" {
			username := parts[1]

			client := updateMutedByUsername(username, false)

			if client != nil {
				fmt.Printf("Khoa mom thanh cong %v\n", username)
				sendMessage(nil, conn, "Khoa mom thanh cong "+username)
			} else {
				fmt.Printf("Khoa mom that bai %v\n", username)
				sendMessage(nil, conn, "Khoa mom that bai "+username)
			}
		}
	} else if strings.HasPrefix(message, "/unmute") {
		parts := strings.Fields(message)

		if len(parts) == 2 && parts[0] == "/unmute" {
			username := parts[1]

			client := updateMutedByUsername(username, true)

			if client != nil {
				fmt.Printf("MO mom thanh cong %v\n", username)
				sendMessage(nil, conn, "MO mom thanh cong "+username)
			} else {
				fmt.Printf("MO mom that bai %v\n", username)
				sendMessage(nil, conn, "MO mom that bai "+username)
			}
		}
	} else if strings.HasPrefix(message, "/kick") {
		parts := strings.Fields(message)

		if len(parts) == 2 && parts[0] == "/kick" {
			username := parts[1]

			client := updateKickedByUsername(username, true)

			if client != nil {
				fmt.Printf("KICK thanh cong %v\n", username)
				sendMessage(nil, conn, "KICK thanh cong "+username)
			} else {
				fmt.Printf("KICK that bai %v\n", username)
				sendMessage(nil, conn, "KICK that bai "+username)
			}
		}
	}
}

func updateMutedByUsername(username string, isMuted bool) *model.Client {
	if username == "" {
		return nil
	}

	for _, client := range clients {
		if username == client.Username {
			client.Muted = isMuted
			return client
		}
	}
	return nil
}

func updateKickedByUsername(username string, isKicked bool) *model.Client {
	if username == "" {
		return nil
	}

	for _, client := range clients {
		if username == client.Username {
			client.Kicked = isKicked
			return client
		}
	}
	return nil
}

func getClientOnline() string {
	msg := ""

	for _, client := range clients {
		if client.Username == "" {
			continue
		}

		msg += client.Username + "\n"
	}

	return msg
}

func findClientByUsername(username string) *model.Client {
	if username == "" {
		return nil
	}

	mul.RLock()
	defer mul.RUnlock()

	for _, client := range clients {
		if client.Username == username {
			return client
		}
	}

	return nil
}

func findClientByConn(conn net.Conn) *model.Client {
	if conn == nil {
		return nil
	}

	for _, client := range clients {
		if conn == client.Conn {
			return client
		}
	}

	return nil
}

func sendMessage(sender net.Conn, receive net.Conn, message string) {

	if receive != nil {
		_, err := receive.Write([]byte(message))

		if err != nil {
			fmt.Printf("Loi gui du lieu %v\n", err)
		}

		fmt.Printf("Gui thanh cong message %v\n", message)
		return
	}

	for _, client := range clients {
		if sender == client.Conn {
			continue
		}

		_, err := client.Conn.Write([]byte(message))

		if err != nil {
			fmt.Printf("Loi gui du lieu %v\n", err)
		}

		fmt.Printf("Gui thanh cong message %v\n", message)
	}
}
