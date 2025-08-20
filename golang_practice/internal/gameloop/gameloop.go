package gameloop

import (
	"golang_practice/pkg/packets/action"
	"golang_practice/pkg/packets/game"
	"golang_practice/pkg/packets/match"
	"golang_practice/pkg/packets/shared"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

const TICK_RATE = 60
const TICK_DURATION = time.Second / TICK_RATE

var (
	pendingJoinRequest = make(map[uint32]*match.JoinMatchRequest)
	pendingActions     = make(map[uint32]*action.Player_Action)
	pendingActionsLock sync.Mutex

	players     = make(map[uint32]*game.PlayerState)
	playersLock sync.RWMutex

	results     = make(map[uint32]*game.GameStateUpdate)
	resultsLock sync.Mutex

	clientAddrs     = make(map[uint32]*net.UDPAddr)
	clientAddrsLock sync.Mutex

	serverConn *net.UDPConn

	speed = 5.0
)

var currentTick int32 = 0

func ApplyPlayerAction(playerId uint32, payload *action.Player_Action, addr *net.UDPAddr) {
	onPlayerJoin(playerId)

	pendingActionsLock.Lock()
	pendingActions[playerId] = payload
	pendingActionsLock.Unlock()

	clientAddrsLock.Lock()
	clientAddrs[playerId] = addr
	clientAddrsLock.Unlock()
}

func Start(conn *net.UDPConn) {
	serverConn = conn

	ticker := time.NewTicker(TICK_DURATION)

	defer ticker.Stop()

	for range ticker.C {
		processIncomingActions()
		updateGameState()
		sendStateToClients()
	}
}

func onPlayerJoin(playerId uint32) {
	playersLock.RLock()
	_, exists := players[playerId]
	playersLock.RUnlock()

	if !exists {
		playersLock.Lock()
		players[playerId] = &game.PlayerState{
			PlayerId: playerId,
			X:        0,
			Y:        0,
			Z:        0,
		}
		playersLock.Unlock()
	}
}

func processIncomingActions() {
	pendingActionsLock.Lock()
	defer pendingActionsLock.Unlock()

	playersLock.Lock()
	defer playersLock.Unlock()

	for playerId, act := range pendingActions {

		playerState, exists := players[playerId]

		if !exists {
			playerState = &game.PlayerState{
				PlayerId: playerId,
			}

			players[playerId] = playerState
		}

		playerState.DirX = act.DirX
		playerState.DirY = act.DirY

		playerState.MoveX = act.MoveX
		playerState.MoveY = act.MoveY

		switch combat := act.CombatAction.(type) {
		case *action.Player_Action_IsAttacking:
			playerState.CombatState = &game.PlayerState_IsAttacking{
				IsAttacking: combat.IsAttacking,
			}
		case *action.Player_Action_IsUsingUlti:
			playerState.CombatState = &game.PlayerState_IsUsingUlti{
				IsUsingUlti: combat.IsUsingUlti,
			}
		case *action.Player_Action_SpellCast:
			playerState.CombatState = &game.PlayerState_SpellId{
				SpellId: combat.SpellCast.SpellId,
			}
		default:
			playerState.CombatState = nil
		}

		playerState.Tick = act.Tick
	}

	pendingActions = make(map[uint32]*action.Player_Action)
}

func updateGameState() {
	playersLock.RLock()
	defer playersLock.RUnlock()

	resultsLock.Lock()
	defer resultsLock.Unlock()

	for _, player := range players {
		currentTick++
		updateMovement(player)

		copied := *player
		results[player.PlayerId] = &game.GameStateUpdate{
			Players: []*game.PlayerState{&copied},
			Tick:    currentTick,
		}
	}
}

func sendStateToClients() {
	resultsLock.Lock()
	defer resultsLock.Unlock()

	clientAddrsLock.Lock()
	defer clientAddrsLock.Unlock()

	for playerId, payload := range results {
		data, err := proto.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal GameStateUpdate for player: %v", err)
			continue
		}

		env := &shared.Envelope{
			Type:    shared.PacketType_GAME_STATE_UPDATE,
			Payload: data,
		}

		packet, err := proto.Marshal(env)

		if err != nil {
			log.Printf("Failed to marshal envelope for player %d: %v", playerId, err)
			continue
		}

		addr, ok := clientAddrs[playerId]
		if !ok {
			log.Printf("No UDP address for player %d", playerId)
			continue
		}

		_, err = serverConn.WriteToUDP(packet, addr)
		if err != nil {
			log.Printf("Failed to send packet to player %d: %v", playerId, err)
		}
	}
}

func updateMovement(player *game.PlayerState) {
	deltaTime := float32(TICK_DURATION.Seconds())
	player.X += player.MoveX * float32(speed) * deltaTime
	player.Y += player.MoveY * float32(speed) * deltaTime
}

func updateCombat(player *game.PlayerState) {

}

func checkCollisions(player *game.PlayerState) {

}
