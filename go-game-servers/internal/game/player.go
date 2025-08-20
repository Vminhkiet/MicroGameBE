package game

import (
	"math"
	"net"

	pb "github.com/Vminhkiet/BattleGround-backend/go-game-server/pkg/protos/game"
)

// Player đại diện cho trạng thái của một người chơi trong game.
type Player struct {
	ID                     string
	CharacterID            string
	ClientAddr             *net.UDPAddr
	Position               *pb.Vector2 // Sử dụng con trỏ tới pb.Vector2 từ Protobuf
	Direction              *pb.Vector2 // Sử dụng con trỏ tới pb.Vector2 từ Protobuf
	Velocity               *pb.Vector2 // Sử dụng con trỏ tới pb.Vector2 từ Protobuf
	Health                 float32
	MaxHealth              float32
	State                  pb.PlayerStateEnum // Sử dụng enum PlayerStateEnum từ Protobuf
	LastProcessedInputTick int32
	Score                  int32
	Kills                  int32
	Deaths                 int32
	// Thêm các thuộc tính khác của người chơi như cooldowns, ammo, v.v.
}

// NewPlayer là hàm khởi tạo một đối tượng Player mới.
// Nó nhận vào ID người chơi, ID nhân vật, vị trí ban đầu (dưới dạng *pb.Vector2) và địa chỉ client.
// Hàm này trả về một con trỏ tới một Player (*Player).
func NewPlayer(id string, charID string, initialPos *pb.Vector2, clientAddr *net.UDPAddr) *Player {
	// Trả về một con trỏ tới một Player struct đã được khởi tạo
	return &Player{
		ID:                     id,
		CharacterID:            charID,
		ClientAddr:             clientAddr,
		Position:               initialPos,
		Direction:              &pb.Vector2{X: 0, Y: 1}, // Mặc định hướng lên (VD: trục Y dương)
		Velocity:               &pb.Vector2{X: 0, Y: 0}, // Ban đầu chưa di chuyển
		Health:                 100.0,
		MaxHealth:              100.0,
		State:                  pb.PlayerStateEnum_IDLE, // Trạng thái ban đầu từ Protobuf enum
		LastProcessedInputTick: 0,                       // Bắt đầu từ tick 0
		Score:                  0,
		Kills:                  0,
		Deaths:                 0,
	}
}

// ApplyInput xử lý một PlayerInput nhận được từ client.
// 'p *Player' là "receiver" của phương thức này. Nó cho biết phương thức này
// hoạt động trên một đối tượng Player cụ thể (và có thể thay đổi nó vì dùng con trỏ '*').
func (p *Player) ApplyInput(input *pb.PlayerInput, currentTick int32) {
	// Logic đơn giản: Nếu client gửi input di chuyển (các trục khác 0)
	// Lưu ý: Protobuf tự động chuyển đổi snake_case (horizontal_axis) sang CamelCase (HorizontalAxis) trong Go
	if input.HorizontalAxis != 0 || input.VerticalAxis != 0 {
		// Cập nhật hướng dựa trên input
		p.Direction.X = input.HorizontalAxis
		p.Direction.Y = input.VerticalAxis

		// Chuẩn hóa vector hướng (rất quan trọng để đảm bảo tốc độ không đổi)
		magnitude := p.Direction.X*p.Direction.X + p.Direction.Y*p.Direction.Y
		if magnitude > 0 {
			magSqrt := float32(math.Sqrt(float64(magnitude))) // Cần import "math"
			p.Direction.X /= magSqrt
			p.Direction.Y /= magSqrt
		} else {
			// Nếu không có hướng (input = 0,0), thì không di chuyển
			p.State = pb.PlayerStateEnum_IDLE // Sử dụng enum từ Protobuf
			p.Velocity.X = 0                  // Đặt vận tốc về 0
			p.Velocity.Y = 0
			return // Không cần xử lý thêm input di chuyển nếu không có hướng
		}

		p.State = pb.PlayerStateEnum_MOVING // Đặt trạng thái là đang di chuyển từ Protobuf enum
	} else {
		p.State = pb.PlayerStateEnum_IDLE // Nếu không có input di chuyển, coi là đứng yên từ Protobuf enum
	}

	if input.GetShootPressed() { // Đối với bool field, Protobuf sinh ra Get<FieldName> (GetShootPressed)
		// Ở đây bạn sẽ thêm logic bắn súng, kiểm tra cooldown...
		p.State = pb.PlayerStateEnum_ATTACKING // Đặt trạng thái tấn công từ Protobuf enum
	} else {
		// Nếu không nhấn Shoot, và không di chuyển, thì trở lại Idle
		if p.State == pb.PlayerStateEnum_ATTACKING { // Chỉ reset nếu trạng thái hiện tại là tấn công
			p.State = pb.PlayerStateEnum_IDLE // Sử dụng enum từ Protobuf
		}
	}

	// Ghi lại tick mà input này đã được xử lý trên server.
	p.LastProcessedInputTick = currentTick
}

// Update cập nhật trạng thái của người chơi qua thời gian.
// Được gọi trong vòng lặp game của Session.
func (p *Player) Update(deltaTime float32) {
	// Nếu đang di chuyển, cập nhật vị trí
	if p.State == pb.PlayerStateEnum_MOVING { // Sử dụng enum từ Protobuf
		speed := float32(5.0) // Tốc độ di chuyển cứng (ví dụ 5 đơn vị/giây)
		p.Position.X += p.Direction.X * speed * deltaTime
		p.Position.Y += p.Direction.Y * speed * deltaTime
	}
	// Giả định nếu tấn công thì quay về idle sau 1 thời gian ngắn (ví dụ đơn giản)
	if p.State == pb.PlayerStateEnum_ATTACKING { // Sử dụng enum từ Protobuf
		// Đây chỉ là ví dụ đơn giản, thực tế cần animation time, cooldown...
		// time.AfterFunc sẽ tạo một goroutine mới, không phù hợp cho game loop
		// Thay vào đó, bạn sẽ dùng một timer trong Player struct và giảm dần nó trong Update
		// Ví dụ: p.AttackTimer -= deltaTime; if p.AttackTimer <= 0 { p.State = PlayerStateEnum_IDLE }
		// Để giữ đơn giản cho người mới bắt đầu, chúng ta tạm thời bỏ qua timer phức tạp ở đây.
		// Trạng thái tấn công sẽ được reset bởi input mới hoặc logic khác.
	}

	// ... Các logic cập nhật khác như hồi máu, hiệu ứng, v.v.
}

// TakeDamage giảm máu của người chơi.
func (p *Player) TakeDamage(amount float32) {
	p.Health -= amount
	if p.Health <= 0 {
		p.Health = 0
		p.State = pb.PlayerStateEnum_DEAD // Đặt trạng thái là đã chết từ Protobuf enum
		// Bạn có thể kích hoạt một sự kiện ở đây để thông báo người chơi đã chết
	}
}

// ToProto chuyển đổi đối tượng Player sang định dạng Protobuf PlayerState.
func (p *Player) ToProto() *pb.PlayerState {
	return &pb.PlayerState{
		UserId:        p.ID,
		CharacterId:   p.CharacterID,
		Position:      p.Position,  // Đã là *pb.Vector2
		Direction:     p.Direction, // Đã là *pb.Vector2
		Health:        p.Health,
		MaxHealth:     p.MaxHealth,
		State:         p.State, // Đã là pb.PlayerStateEnum
		LastInputTick: p.LastProcessedInputTick,
		Score:         p.Score,
		Kills:         p.Kills,
		Deaths:        p.Deaths,
	}
}
