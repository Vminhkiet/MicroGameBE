package main

import (
	"fmt"
	"golang_practice/internal/gameloop"
	"golang_practice/internal/handler"
	"golang_practice/internal/router"
	"golang_practice/pkg/packets/shared"
	"net"
)

func main() {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 9999,
	})
	if err != nil {
		fmt.Printf("Lỗi mở server: %v\n", err)
		return
	}

	go gameloop.Start(conn)

	router.Register(shared.PacketType_PING_REQUEST, handler.HandlePingRequest)
	router.Register(shared.PacketType_PLAYER_ACTION, handler.HandlePlayerAction)

	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Lỗi đọc UDP: %v\n", err)
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		go router.RoutePacket(conn, data, clientAddr)
	}
}
