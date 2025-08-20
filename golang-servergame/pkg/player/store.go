package player

import (
	"net"
	"sync"
	"time"
)

var (
	players  = make(map[uint32]*Player)
	addrToID = make(map[string]uint32)
	mu       sync.RWMutex
)

func UpdateorCreate(addr *net.UDPAddr, id uint32, nickname string) *Player {
	mu.Lock()
	defer mu.Unlock()

	p, exists := players[id]

	if !exists {
		p = &Player{
			ID:       id,
			Nickname: nickname,
			Addr:     addr,
		}

		players[id] = p
	}

	addrToID[addr.String()] = id
	p.LastSeen = time.Now().Unix()

	return p
}

func GetByAddr(addr string) (*Player, bool) {
	mu.RLock()
	defer mu.RUnlock()

	id, ok := addrToID[addr]

	if !ok {
		return nil, false
	}

	p, ok := players[id]

	return p, ok
}

func UpdateLastSeen(id uint32) {
	mu.Lock()
	defer mu.Unlock()

	if p, ok := players[id]; ok {
		p.LastSeen = time.Now().Unix()
	}
}

func GetByID(id uint32) *Player {
	mu.RLock()
	defer mu.RUnlock()

	return players[id]
}

func GetAll() []*Player {
	mu.Lock()
	defer mu.Unlock()

	result := make([]*Player, 0, len(players))

	for _, p := range players {
		result = append(result, p)
	}

	return result
}

func Remove(id uint32) {
	mu.Lock()
	defer mu.Unlock()
	delete(players, id)
}

func ForEach(f func(*Player)) {
	mu.RLock()
	defer mu.RUnlock()

	for _, p := range players {
		f(p)
	}
}
