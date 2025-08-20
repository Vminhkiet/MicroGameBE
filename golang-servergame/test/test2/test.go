package main

import (
	"log"
	"net"
	"time"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/action"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/game"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/match"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
	"google.golang.org/protobuf/proto"
)

func main() {
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9999,
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal("Không kết nối được:", err)
	}
	defer conn.Close()

	// 1. Gửi JoinMatchRequest
	join := &match.JoinMatchRequest{
		PlayerId: 123,
		Nickname: "TestPlayer",
	}
	sendPacket(conn, shared.PacketType_JOIN_MATCH_REQUEST, join)

	// 2. Gửi PlayerAction mỗi 200ms
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			actionMsg := &action.PlayerAction{
				PlayerId: 123,
				X:        float32(time.Now().UnixNano()%100) / 10.0,
				Y:        float32(time.Now().UnixNano()%100) / 10.0,
				MoveX:    1,
				MoveY:    0,
				Tick:     int32(time.Now().UnixNano() / 1e6),
			}
			sendPacket(conn, shared.PacketType_PLAYER_ACTION, actionMsg)
		}
	}()

	// 3. Nhận dữ liệu từ server
	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Không nhận được:", err)
			continue
		}

		var env shared.Envelope
		if err := proto.Unmarshal(buf[:n], &env); err != nil {
			log.Println("Lỗi giải mã Envelope:", err)
			continue
		}

		switch env.Type {
		case shared.PacketType_GAME_STATE_UPDATE:
			var state game.GameStateUpdate // hoặc game.GameStateUpdate nếu bạn import đúng
			if err := proto.Unmarshal(env.Payload, &state); err != nil {
				log.Println("Lỗi giải mã GameStateUpdate:", err)
				continue
			}
			log.Printf("🎮 Tick: %d", state.Tick)
			for _, p := range state.Players {
				log.Printf("👤 Player %d at (%.2f, %.2f)", p.PlayerId, p.X, p.Y)
			}
		default:
			log.Println("📦 Nhận gói không xác định:", env.Type)
		}
	}
}

// Gửi bất kỳ message nào dưới dạng Envelope
func sendPacket(conn *net.UDPConn, packetType shared.PacketType, msg proto.Message) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		log.Println("❌ Marshal lỗi:", err)
		return
	}

	env := &shared.Envelope{
		Type:    packetType,
		Payload: payload,
	}
	packetData, err := proto.Marshal(env)
	if err != nil {
		log.Println("❌ Envelope lỗi:", err)
		return
	}

	conn.Write(packetData)
}
