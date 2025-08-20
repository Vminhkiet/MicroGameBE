package main

import (
	"fmt"
	"log"
	"net"
	"strings" // Thêm import này để sử dụng strings.Contains
	"time"

	pb "github.com/Vminhkiet/BattleGround-backend/go-game-server/pkg/protos/game" // Đảm bảo đường dẫn này đúng
	"google.golang.org/protobuf/proto"
)

const (
	serverPort = 8080
	serverIP   = "127.0.0.1" // Địa chỉ IP của server (localhost)
	playerID   = "test_player_1"
	sessionID  = "default_session" // Phải khớp với sessionID trên server
	tickRate   = 60                // Tốc độ tick của client (để gửi input)
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		log.Fatalf("Lỗi phân giải địa chỉ server: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("Lỗi kết nối tới server: %v", err)
	}
	defer conn.Close()

	log.Printf("Đã kết nối tới server UDP tại %s", serverAddr.String())

	// Goroutine để lắng nghe GameStateUpdate từ server
	go listenForGameState(conn)

	// Gửi JoinRequest
	sendJoinRequest(conn, playerID, sessionID, "avatar_01")
	time.Sleep(1 * time.Second) // Đợi server xử lý

	// Gửi PlayerInput liên tục
	ticker := time.NewTicker(time.Second / time.Duration(tickRate))
	defer ticker.Stop()

	log.Println("Bắt đầu gửi PlayerInput...")
	for i := 0; i < 300; i++ { // Gửi input trong khoảng 5 giây (300 ticks / 60 ticks/s)
		<-ticker.C
		sendPlayerInput(conn, playerID, 0.5, 0.5, false, false) // Di chuyển chéo
	}

	// Gửi LeaveRequest
	sendLeaveRequest(conn, playerID, sessionID)
	time.Sleep(1 * time.Second) // Đợi server xử lý

	log.Println("Client đã hoàn thành.")
}

// sendJoinRequest gửi thông điệp JoinRequest tới server.
func sendJoinRequest(conn *net.UDPConn, pID, sID, charID string) {
	joinReq := &pb.JoinRequest{
		PlayerId:    pID,
		SessionId:   sID,
		CharacterId: charID,
	}
	clientMsg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_JoinRequest{
			JoinRequest: joinReq,
		},
	}
	sendClientMessage(conn, clientMsg, "JoinRequest")
}

// sendPlayerInput gửi thông điệp PlayerInput tới server.
func sendPlayerInput(conn *net.UDPConn, pID string, hAxis, vAxis float32, jump, shoot bool) {
	input := &pb.PlayerInput{
		PlayerId:       pID,
		HorizontalAxis: hAxis,
		VerticalAxis:   vAxis,
		JumpPressed:    jump,
		ShootPressed:   shoot,
	}
	clientMsg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_PlayerInput{
			PlayerInput: input,
		},
	}
	sendClientMessage(conn, clientMsg, "PlayerInput")
}

// sendLeaveRequest gửi thông điệp LeaveRequest tới server.
func sendLeaveRequest(conn *net.UDPConn, pID, sID string) {
	leaveReq := &pb.LeaveRequest{
		PlayerId:  pID,
		SessionId: sID,
	}
	clientMsg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_LeaveRequest{
			LeaveRequest: leaveReq,
		},
	}
	sendClientMessage(conn, clientMsg, "LeaveRequest")
}

// sendClientMessage mã hóa và gửi ClientMessage tới server.
func sendClientMessage(conn *net.UDPConn, msg *pb.ClientMessage, msgType string) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Lỗi mã hóa %s: %v", msgType, err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Lỗi gửi %s: %v", msgType, err)
	}
	// log.Printf("Đã gửi %s", msgType)
}

// listenForGameState lắng nghe và giải mã GameStateUpdate từ server.
func listenForGameState(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		// Đặt timeout để không bị chặn vĩnh viễn
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout, thử lại
			}
			// Kiểm tra lỗi khi kết nối bị đóng
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("Kết nối client đã đóng, dừng lắng nghe GameState.")
				return
			}
			log.Printf("Lỗi đọc gói tin GameStateUpdate: %v", err)
			continue
		}

		var serverMsg pb.ServerMessage
		err = proto.Unmarshal(buffer[:n], &serverMsg)
		if err != nil {
			log.Printf("Lỗi giải mã thông điệp server: %v", err)
			continue
		}

		switch serverMsg.Payload.(type) {
		case *pb.ServerMessage_GameStateUpdate:
			gameState := serverMsg.GetGameStateUpdate()
			// log.Printf("Đã nhận GameStateUpdate (Tick: %d) với %d người chơi.",
			// 	gameState.GetServerTick(), len(gameState.GetPlayersStates()))

			// In chi tiết trạng thái người chơi đầu tiên (nếu có)
			if len(gameState.GetPlayersStates()) > 0 {
				pState := gameState.GetPlayersStates()[0]
				log.Printf("  -> Player %s: Pos=(%.2f, %.2f), Health=%.0f, State=%s, LastInputTick=%d",
					pState.GetUserId(), pState.GetPosition().GetX(), pState.GetPosition().GetY(),
					pState.GetHealth(), pState.GetState().String(), pState.GetLastInputTick())
			}
		default:
			log.Printf("Đã nhận loại thông điệp server không xác định.")
		}
	}
}
