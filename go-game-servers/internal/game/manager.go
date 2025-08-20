package game

import (
	"errors"
	"log"
	"net"
	"sync"

	pb "github.com/Vminhkiet/BattleGround-backend/go-game-server/pkg/protos/game"
)

// GameManager quản lý tất cả các phiên game đang hoạt động.
// Nó chịu trách nhiệm tạo, truy xuất và đóng các phiên game.
type GameManager struct {
	sessions map[string]*Session // Map sessionID to Session
	mu       sync.RWMutex        // Mutex để bảo vệ truy cập đồng thời vào sessions map

	// sendGameStateFunc là một hàm callback được cung cấp bởi tầng cao hơn (ví dụ: UDPServer)
	// để gửi GameStateUpdate đến một client cụ thể.
	// Tham số: sessionID, GameStateUpdate, playerID (để xác định địa chỉ gửi)
	sendGameStateFunc func(sessionID string, state *pb.GameStateUpdate, playerID string)
}

// NewGameManager tạo một GameManager mới.
// Tham số `sendGameStateFunc` là một hàm mà GameManager sẽ sử dụng để gửi dữ liệu trạng thái game
// trở lại các client thông qua tầng mạng.
func NewGameManager(sendGameStateFunc func(sessionID string, state *pb.GameStateUpdate, playerID string)) *GameManager {
	return &GameManager{
		sessions:          make(map[string]*Session),
		sendGameStateFunc: sendGameStateFunc,
	}
}

// CreateSession tạo một phiên game mới với ID và tốc độ tick đã cho.
// Nếu phiên đã tồn tại, nó sẽ trả về lỗi.
func (gm *GameManager) CreateSession(sessionID string, tickRate int) (*Session, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.sessions[sessionID]; exists {
		log.Printf("Session %s already exists.", sessionID)
		return nil, errors.New("session already exists")
	}

	newSession := NewSession(sessionID, tickRate)
	gm.sessions[sessionID] = newSession

	// Khởi động vòng lặp game của phiên trong một goroutine riêng biệt.
	// Hàm callback được truyền vào sẽ được phiên gọi mỗi khi có cập nhật trạng thái.
	go newSession.Start(func(sID string, state *pb.GameStateUpdate) {
		// Hàm ẩn danh này sẽ được Session gọi để thông báo trạng thái game đã sẵn sàng gửi.
		// GameManager sau đó sẽ gửi trạng thái này đến tất cả người chơi trong phiên.
		// Dùng RLock để đọc danh sách người chơi một cách an toàn.
		newSession.mu.RLock()
		playersInSession := make([]*Player, 0, len(newSession.Players))
		for _, p := range newSession.Players {
			playersInSession = append(playersInSession, p)
		}
		newSession.mu.RUnlock()

		for _, player := range playersInSession {
			// Gọi hàm gửi trạng thái được cung cấp bởi tầng mạng (ví dụ: UDPServer).
			gm.sendGameStateFunc(sID, state, player.ID)
		}
	})

	log.Printf("Phiên %s đã được tạo và bắt đầu.", sessionID)
	return newSession, nil
}

// GetSession trả về một phiên game dựa trên ID của nó.
// Trả về nil nếu phiên không tồn tại.
func (gm *GameManager) GetSession(sessionID string) *Session {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.sessions[sessionID]
}

// AddPlayerToSession thêm một người chơi vào một phiên game cụ thể.
// Nếu phiên không tồn tại, nó sẽ tạo một phiên mới với tốc độ tick mặc định là 60.
// Nếu người chơi đã tồn tại trong phiên, nó sẽ cập nhật địa chỉ client của người chơi.
func (gm *GameManager) AddPlayerToSession(sessionID string, playerID string, clientAddr *net.UDPAddr) (*Player, *Session, error) {
	gm.mu.RLock()
	session, exists := gm.sessions[sessionID]
	gm.mu.RUnlock()

	if !exists {
		// Nếu phiên không tồn tại, tạo một phiên mới.
		// Trong một game thực tế, bạn có thể muốn có một hệ thống sảnh chờ riêng
		// hoặc trả về lỗi nếu phiên không được tạo trước.
		log.Printf("Phiên %s không tồn tại, đang tạo một phiên mới cho người chơi %s.", sessionID, playerID)
		var err error
		session, err = gm.CreateSession(sessionID, 60) // Tốc độ tick mặc định 60
		if err != nil {
			log.Printf("Không thể tạo phiên %s: %v", sessionID, err)
			return nil, nil, err
		}
	}

	// Kiểm tra xem người chơi đã tồn tại trong phiên chưa.
	session.mu.RLock() // Dùng RLock để đọc map Players
	player, playerExistsInSession := session.Players[playerID]
	session.mu.RUnlock()

	if playerExistsInSession {
		log.Printf("Người chơi %s đã có trong phiên %s. Đang cập nhật địa chỉ client nếu thay đổi.", playerID, sessionID)
		// Cập nhật địa chỉ client nếu nó thay đổi (ví dụ: client reconnect với port mới).
		session.mu.Lock() // Dùng Lock để ghi vào Player struct
		if player.ClientAddr.String() != clientAddr.String() {
			log.Printf("Đang cập nhật địa chỉ client cho người chơi %s từ %s sang %s", playerID, player.ClientAddr.String(), clientAddr.String())
			player.ClientAddr = clientAddr
		}
		session.mu.Unlock()
	} else {
		// Tạo người chơi mới và thêm vào phiên.
		// Sử dụng các giá trị mặc định cho charID và initialPos.
		// Đã sửa lỗi: Thêm dấu '&' để truyền con trỏ tới pb.Vector2
		player = NewPlayer(playerID, "default_character", &pb.Vector2{X: 0, Y: 0}, clientAddr)
		session.AddPlayer(player)
	}

	return player, session, nil
}

// RemovePlayerFromSession xóa một người chơi khỏi một phiên game.
// Nếu phiên trở nên trống rỗng sau khi người chơi rời đi, phiên sẽ bị dừng và xóa khỏi GameManager.
func (gm *GameManager) RemovePlayerFromSession(sessionID string, playerID string) {
	gm.mu.RLock()
	session, exists := gm.sessions[sessionID]
	gm.mu.RUnlock()

	if !exists {
		log.Printf("Không thể xóa người chơi %s: Phiên %s không tồn tại.", playerID, sessionID)
		return
	}

	session.RemovePlayer(playerID)

	// Nếu phiên trở nên trống rỗng sau khi người chơi rời đi, dừng và xóa nó khỏi manager.
	session.mu.RLock()
	numPlayers := len(session.Players)
	session.mu.RUnlock()

	if numPlayers == 0 {
		gm.mu.Lock()
		delete(gm.sessions, sessionID)
		gm.mu.Unlock()
		session.Stop() // Dừng vòng lặp game của phiên
		log.Printf("Phiên %s trống và đã bị xóa khỏi GameManager.", sessionID)
	}
}

// HandlePlayerInput nhận input từ một người chơi và chuyển tiếp nó đến phiên game tương ứng.
// Hàm này thường được gọi bởi tầng mạng (ví dụ: UDPServer) khi nhận được gói tin input.
func (gm *GameManager) HandlePlayerInput(sessionID string, playerID string, input *pb.PlayerInput, clientAddr *net.UDPAddr) {
	gm.mu.RLock()
	session, exists := gm.sessions[sessionID]
	gm.mu.RUnlock()

	if !exists {
		log.Printf("Đã nhận đầu vào cho phiên %s không tồn tại. ID người chơi: %s", sessionID, playerID)
		// Tùy chọn: có thể thử thêm người chơi vào một phiên nếu đây là một nỗ lực kết nối mới.
		// Hiện tại, chỉ ghi log và trả về.
		return
	}

	session.ApplyPlayerInput(playerID, input, clientAddr)
}
