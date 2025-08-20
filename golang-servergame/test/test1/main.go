package main

import (
	"log"
	"net"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/match"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
	"google.golang.org/protobuf/proto"
)

func main() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9999,
	})

	if err != nil {
		log.Fatal("❌ Không thể kết nối đến server:", err)
	}

	defer conn.Close()

	payload := &match.JoinMatchRequest{
		PlayerId: 123,
		Nickname: "TestPlayer",
	}

	data, err := proto.Marshal(payload)

	if err != nil {
		log.Fatal("❌ Marshal JoinMatchRequest lỗi:", err)
	}

	env := &shared.Envelope{
		Type:    shared.PacketType_JOIN_MATCH_REQUEST,
		Payload: data,
	}

	packet, err := proto.Marshal(env)

	if err != nil {
		log.Fatal("❌ Marshal envelope lỗi:", err)
	}

	_, err = conn.Write(packet)

	if err != nil {
		log.Fatal("❌ Gửi gói tin lỗi:", err)
	}

	log.Println("✅ Đã gửi JoinMatchRequest")
}
