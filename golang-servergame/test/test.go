package main

import (
	"log"
	"net"
	"time"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/ping"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
	"google.golang.org/protobuf/proto"
)

func main() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9999,
	})

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	pingReq := &ping.PingRequest{
		ClientTime: time.Now().UnixMilli(),
	}

	payload, _ := proto.Marshal(pingReq)

	env := &shared.Envelope{
		Type:    shared.PacketType_PING_REQUEST,
		Payload: payload,
	}

	packetData, err := proto.Marshal(env)

	if err != nil {
		log.Fatalf("❌ Failed to marshal Envelope: %v", err)
	}

	_, err = conn.Write(packetData)

	if err != nil {
		log.Fatalf("❌ Failed to send packet: %v", err)
	}

	buf := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	n, _, err := conn.ReadFromUDP(buf)

	if err != nil {
		log.Println("⚠️ No response from server:", err)
		return
	}

	var response shared.Envelope
	if err := proto.Unmarshal(buf[:n], &response); err != nil {
		log.Println("⚠️ Failed to unmarshal response:", err)
		return
	}

	log.Printf("📨 Received response packet type: %v", response.Type)
}
