package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/VMinhKiet/golang-servergame/client"
	"github.com/VMinhKiet/golang-servergame/internal/gameloop.go"
	"github.com/VMinhKiet/golang-servergame/internal/handler"
	"github.com/VMinhKiet/golang-servergame/internal/router"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
)

var ()

func main() {

	router.Register(shared.PacketType_PLAYER_ACTION, handler.HandlePlayerAction)
	router.Register(shared.PacketType_PING_REQUEST, handler.HandlePing)
	router.Register(shared.PacketType_JOIN_MATCH_REQUEST, handler.HandleJoinMatch)

	StartClientCleanUp(10*time.Second, 30*time.Second)

	addr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		fmt.Printf("Loi lang nghe client %v", err)
		return
	}

	defer conn.Close()

	go gameloop.Start(conn)

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			log.Printf("❌ Lỗi đọc packet: %v", err)
			continue
		}

		c := client.Update(clientAddr)

		log.Println("Client:", c.Addr.String())

		go router.RoutePacket(buffer[:n], clientAddr.String())
	}
}

func StartClientCleanUp(interval time.Duration, timeout time.Duration) {
	go func() {
		for {
			time.Sleep(interval)

			client.ForEach(func(c *client.Client) {
				if time.Since(c.LastSeen) > timeout {
					log.Println("🔴 Removing inactive client:", c.Addr)
					client.Remove(c.Addr)
				}
			})
		}
	}()
}
