package game

import (
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/Vminhkiet/BattleGround-backend/go-game-server/pkg/protos/game"
)

type Session struct {
	ID      string
	Players map[string]*Player // Map userID to Player
	// InputsQueue map[string][]PlayerInput // Hàng đợi input cho mỗi người chơi
	// Để đơn giản, chúng ta sẽ xử lý input ngay lập tức, nhưng trong game lớn hơn, cần hàng đợi.

	tickRate    int           // Số tick mỗi giây (ví dụ: 60)
	currentTick int32         // Tick hiện tại của server
	stopChan    chan struct{} // Kênh để dừng game loop
	mu          sync.RWMutex  // Mutex để bảo vệ truy cập đồng thời vào Players map và các dữ liệu khác
}

func NewSession(id string, tickRate int) *Session {
	return &Session{
		ID:          id,
		Players:     make(map[string]*Player),
		tickRate:    tickRate,
		currentTick: 0,
		stopChan:    make(chan struct{}),
		mu:          sync.RWMutex{},
	}
}

func (s *Session) AddPlayer(player *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Players[player.ID] = player
	log.Printf("Player %s (%s) joined session %s", player.ID, player.ClientAddr.String(), s.ID)
}

func (s *Session) RemovePlayer(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Players, playerID)
	log.Printf("Player %s left session %s", playerID, s.ID)
	// Nếu không còn người chơi nào, có thể đóng session
	if len(s.Players) == 0 {
		log.Printf("Session %s is empty, stopping.", s.ID)
		s.Stop()
	}
}

func (s *Session) Start(sendGameState func(sessionID string, state *pb.GameStateUpdate)) {
	log.Printf("Session %s started with tick rate %d", s.ID, s.tickRate)
	ticker := time.NewTicker(time.Second / time.Duration(s.tickRate))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.currentTick++
			s.mu.RLock() // Dùng RLock khi đọc Player data

			// 1. Xử lý Input (trong vòng lặp này, inputs sẽ được xử lý khi nhận)
			// (Giả định UDPServer sẽ gọi s.ApplyPlayerInput ngay khi nhận được)

			// 2. Cập nhật trạng thái của tất cả người chơi
			deltaTime := float32(1.0 / float32(s.tickRate))
			playerStatesForUpdate := make([]*pb.PlayerState, 0, len(s.Players))

			for _, p := range s.Players {
				p.Update(deltaTime)                                                // Cập nhật vị trí, trạng thái, v.v.
				playerStatesForUpdate = append(playerStatesForUpdate, p.ToProto()) // Chuyển đổi sang Protobuf
			}
			s.mu.RUnlock()

			// 3. Tạo và gửi GameStateUpdate
			gameStateUpdate := &pb.GameStateUpdate{
				ServerTick: int64(s.currentTick),
				// players_transforms không còn phù hợp nếu muốn gửi full state.
				// Hãy đổi GameStateUpdate trong .proto để dùng repeated PlayerState.
				PlayersStates: playerStatesForUpdate, // <<== Đảm bảo đã sửa game_messages.proto
				// ... thêm các đối tượng game khác
			}

			sendGameState(s.ID, gameStateUpdate)

		case <-s.stopChan:
			log.Printf("Session %s stopped.", s.ID)
			return
		}
	}
}

// Stop dừng vòng lặp game của session.
func (s *Session) Stop() {
	close(s.stopChan)
}

// ApplyPlayerInput áp dụng input từ một người chơi.
// Đây là hàm mà UDPServer sẽ gọi khi nhận được PlayerInput.
func (s *Session) ApplyPlayerInput(playerID string, input *pb.PlayerInput, clientAddr *net.UDPAddr) {
	s.mu.RLock()
	player, exists := s.Players[playerID]
	s.mu.RUnlock()

	if !exists {
		// Người chơi chưa tham gia session hoặc đã rời đi.
		// Có thể thêm logic để thông báo lỗi hoặc bỏ qua input.
		log.Printf("Received input for unknown player %s in session %s", playerID, s.ID)
		return
	}
	// Cập nhật địa chỉ client nếu nó thay đổi (ví dụ: client reconnect với port mới)
	if player.ClientAddr.String() != clientAddr.String() {
		log.Printf("Updating client address for player %s from %s to %s", playerID, player.ClientAddr.String(), clientAddr.String())
		player.ClientAddr = clientAddr
	}

	player.ApplyInput(input, s.currentTick)
}

// GetPlayers trả về danh sách người chơi trong session.
func (s *Session) GetPlayers() map[string]*Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Players
}
