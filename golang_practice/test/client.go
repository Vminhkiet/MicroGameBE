	package main

	import (
		"fmt"
		"golang_practice/pkg/packets/action"
		"golang_practice/pkg/packets/game"
		"golang_practice/pkg/packets/ping"
		"golang_practice/pkg/packets/shared"
		"net"
		"time"

		"google.golang.org/protobuf/proto"
	)

	var (
		tick     int32  = 0
		playerID uint32 = 1325
		serverIP        = "udp-server:9999"
	)

	func main() {
		conn, err := net.DialUDP("udp", nil, resolveAddr(serverIP))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		buf := make([]byte, 2048)

		go receivePackets(buf, conn)

		// Gửi PlayerAction mỗi 1/60 giây
		go func() {
			ticker := time.NewTicker(time.Second / 60)
			defer ticker.Stop()

			for range ticker.C {
				sendPlayerAction(conn)
			}
		}()

		// Gửi Ping mỗi 5 giây
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			sendPing(conn)
		}
	}

	func resolveAddr(addr string) *net.UDPAddr {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			panic(err)
		}
		return udpAddr
	}

	func receivePackets(buf []byte, conn *net.UDPConn) {
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Printf("❌ Lỗi đọc dữ liệu: %v\n", err)
				continue
			}

			var env shared.Envelope
			err = proto.Unmarshal(buf[:n], &env)
			if err != nil {
				fmt.Printf("❌ Lỗi giải mã envelope: %v\n", err)
				continue
			}

			switch env.Type {
			case shared.PacketType_PING_RESPONSE:
				handlePingResponse(env.Payload)

			case shared.PacketType_GAME_STATE_UPDATE:
				handleGameStateUpdate(env.Payload)

			default:
				fmt.Printf("⚠️ Không có handler cho loại packet: %v\n", env.Type)
			}
		}
	}

	func sendPing(conn *net.UDPConn) {
		payload := &ping.PingRequest{
			ClientTime: time.Now().UnixNano(),
		}

		data, _ := proto.Marshal(payload)

		msg := &shared.Envelope{
			Type:    shared.PacketType_PING_REQUEST,
			Payload: data,
		}

		packet, _ := proto.Marshal(msg)
		conn.Write(packet)
	}

	func sendPlayerAction(conn *net.UDPConn) {
		payload := &action.Player_Action{
			PlayerId: playerID,
			MoveX:    1.0,
			MoveY:    0.0,
			DirX:     1.0,
			DirY:     0.0,
			CombatAction: &action.Player_Action_IsAttacking{
				IsAttacking: true,
			},
			Tick: tick,
		}

		tick++

		data, _ := proto.Marshal(payload)

		msg := &shared.Envelope{
			Type:    shared.PacketType_PLAYER_ACTION,
			Payload: data,
		}

		packet, _ := proto.Marshal(msg)
		conn.Write(packet)
	}

	func handlePingResponse(data []byte) {
		var payload ping.PingResponse
		err := proto.Unmarshal(data, &payload)
		if err != nil {
			fmt.Printf("❌ Lỗi giải mã PingResponse: %v\n", err)
			return
		}

		pingMs := (time.Now().UnixNano() - payload.ClientTime) / 1e6
		fmt.Printf("📶 Ping: %d ms (ClientTime: %v, ServerTime: %v)\n", pingMs, payload.ClientTime, payload.ServerTime)
	}

	func handleGameStateUpdate(data []byte) {
		var payload game.GameStateUpdate
		err := proto.Unmarshal(data, &payload)
		if err != nil {
			fmt.Printf("❌ Lỗi giải mã GameStateUpdate: %v\n", err)
			return
		}

		fmt.Printf("🎮 GameStateUpdate Tick %d:\n", payload.Tick)
		for _, player := range payload.Players {
			fmt.Printf("  📍 Player %d | Pos(%.2f, %.2f)\n", player.PlayerId, player.X, player.Y)
		}
	}
