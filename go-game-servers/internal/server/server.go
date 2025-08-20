package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Vminhkiet/BattleGround-backend/go-game-server/internal/game"
	"github.com/Vminhkiet/BattleGround-backend/go-game-server/internal/metrics"
	pb "github.com/Vminhkiet/BattleGround-backend/go-game-server/pkg/protos/game"
	"google.golang.org/protobuf/proto" // Import thư viện Protobuf
)

// UDPServer đại diện cho máy chủ game UDP.
type UDPServer struct {
	Addr        *net.UDPAddr      // Địa chỉ lắng nghe của server
	Conn        *net.UDPConn      // Kết nối UDP
	GameManager *game.GameManager // Quản lý các phiên game
	Clients     sync.Map          // Map để lưu trữ địa chỉ của các client: map[string]*net.UDPAddr (playerID -> UDPAddr)
	mu          sync.RWMutex      // Mutex để bảo vệ truy cập đồng thời vào Clients map
	StopChan    chan struct{}     // Kênh để dừng server
}

// NewUDPServer tạo một thể hiện mới của UDPServer.
// Nó nhận địa chỉ lắng nghe và trả về con trỏ tới UDPServer.
func NewUDPServer(port int) (*UDPServer, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("không thể phân giải địa chỉ UDP: %w", err)
	}

	server := &UDPServer{
		Addr:     addr,
		Clients:  sync.Map{}, // Sử dụng sync.Map cho truy cập đồng thời an toàn
		StopChan: make(chan struct{}),
	}

	// Khởi tạo GameManager và truyền hàm sendGameStateToPlayer của server vào.
	server.GameManager = game.NewGameManager(server.sendGameStateToPlayer)

	return server, nil
}

// Start bắt đầu lắng nghe các gói tin UDP.
func (s *UDPServer) Start() error {
	var err error
	s.Conn, err = net.ListenUDP("udp", s.Addr)
	if err != nil {
		return fmt.Errorf("không thể lắng nghe UDP: %w", err)
	}
	log.Printf("Máy chủ UDP đang lắng nghe trên %s", s.Addr.String())

	// Bắt đầu goroutine để lắng nghe gói tin
	go s.listenForPackets()

	return nil
}

// Stop dừng máy chủ UDP.
func (s *UDPServer) Stop() {
	log.Println("Đang dừng máy chủ UDP...")
	close(s.StopChan) // Gửi tín hiệu dừng đến goroutine lắng nghe
	if s.Conn != nil {
		s.Conn.Close() // Đóng kết nối UDP
	}
	log.Println("Máy chủ UDP đã dừng.")
}

// listenForPackets lắng nghe các gói tin UDP đến.
func (s *UDPServer) listenForPackets() {
	buffer := make([]byte, 1024) // Kích thước buffer cho gói tin UDP

	for {
		select {
		case <-s.StopChan:
			return // Dừng goroutine khi nhận tín hiệu dừng
		default:
			// Đặt timeout để không bị chặn vĩnh viễn và có thể kiểm tra StopChan
			s.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, clientAddr, err := s.Conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout, thử lại
				}
				log.Printf("Lỗi đọc gói tin UDP: %v", err)
				continue
			}

			metrics.PacketCount.Inc()
			// Xử lý gói tin trong một goroutine riêng để không chặn vòng lặp lắng nghe
			go s.handlePacket(buffer[:n], clientAddr)
		}
	}
}

// handlePacket xử lý một gói tin UDP đã nhận.
func (s *UDPServer) handlePacket(data []byte, clientAddr *net.UDPAddr) {
	var clientMsg pb.ClientMessage
	err := proto.Unmarshal(data, &clientMsg)
	if err != nil {
		log.Printf("Lỗi giải mã thông điệp client từ %s: %v", clientAddr.String(), err)
		return
	}

	// Kiểm tra loại payload của thông điệp client
	switch clientMsg.Payload.(type) {
	case *pb.ClientMessage_PlayerInput:
		input := clientMsg.GetPlayerInput()
		playerID := input.GetPlayerId() // Lấy PlayerID từ input
		sessionID := "default_session"  // TODO: Cần một cách để xác định sessionID từ client hoặc logic game

		// Cập nhật địa chỉ client trong Clients map
		s.Clients.Store(playerID, clientAddr)

		// Thêm người chơi vào session nếu chưa có (hoặc cập nhật địa chỉ)
		_, _, err := s.GameManager.AddPlayerToSession(sessionID, playerID, clientAddr)
		if err != nil {
			log.Printf("Lỗi khi thêm/cập nhật người chơi %s vào phiên %s: %v", playerID, sessionID, err)
			return
		}

		// Chuyển tiếp input đến GameManager
		s.GameManager.HandlePlayerInput(sessionID, playerID, input, clientAddr)
		// log.Printf("Đã nhận PlayerInput từ %s (PlayerID: %s) cho phiên %s", clientAddr.String(), playerID, sessionID)

	// TODO: Thêm các loại thông điệp client khác (ví dụ: JoinRequest, LeaveRequest)
	default:
		log.Printf("Đã nhận loại thông điệp client không xác định từ %s", clientAddr.String())
	}
}

// sendGameStateToPlayer là hàm callback được truyền vào GameManager.
// Nó chịu trách nhiệm mã hóa GameStateUpdate và gửi nó đến client tương ứng.
func (s *UDPServer) sendGameStateToPlayer(sessionID string, state *pb.GameStateUpdate, playerID string) {
	// Lấy địa chỉ client từ Clients map
	addrVal, ok := s.Clients.Load(playerID)
	if !ok {
		log.Printf("Không tìm thấy địa chỉ client cho người chơi %s trong phiên %s. Không thể gửi GameStateUpdate.", playerID, sessionID)
		return
	}
	clientAddr, ok := addrVal.(*net.UDPAddr)
	if !ok {
		log.Printf("Địa chỉ client cho người chơi %s không đúng định dạng.", playerID)
		return
	}

	serverMsg := &pb.ServerMessage{
		Payload: &pb.ServerMessage_GameStateUpdate{
			GameStateUpdate: state,
		},
	}

	data, err := proto.Marshal(serverMsg)
	if err != nil {
		log.Printf("Lỗi mã hóa GameStateUpdate cho người chơi %s: %v", playerID, err)
		return
	}

	_, err = s.Conn.WriteToUDP(data, clientAddr)
	if err != nil {
		log.Printf("Lỗi gửi GameStateUpdate đến %s (PlayerID: %s): %v", clientAddr.String(), playerID, err)
	}
	// log.Printf("Đã gửi GameStateUpdate cho người chơi %s (%s) của phiên %s. Tick: %d",
	// 	playerID, clientAddr.String(), sessionID, state.GetServerTick())
}

// Main function (ví dụ về cách sử dụng server)
/*
func main() {
	server, err := NewUDPServer(8080) // Lắng nghe trên cổng 8080
	if err != nil {
		log.Fatalf("Không thể tạo máy chủ UDP: %v", err)
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("Không thể khởi động máy chủ UDP: %v", err)
	}

	// Giữ cho main goroutine chạy để server không thoát ngay lập tức
	select {}
}
*/
