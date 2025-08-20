package client

import (
	"net"
	"sync"
	"time"
)

var (
	clients = make(map[string]*Client)
	mu      sync.RWMutex
)

func Update(addr *net.UDPAddr) *Client {
	key := addr.String()

	mu.Lock()
	defer mu.Unlock()

	c, ok := clients[key]

	if !ok {
		c = &Client{
			Addr: addr,
		}

		clients[key] = c
	}

	c.LastSeen = time.Now()

	return c
}

func Get(addr string) (*Client, bool) {
	mu.RLock()
	defer mu.RUnlock()

	c, ok := clients[addr]
	return c, ok
}

func Remove(addr *net.UDPAddr) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, addr.String())
}

func ForEach(f func(*Client)) {
	mu.RLock()
	defer mu.RUnlock()

	for _, c := range clients {
		f(c)
	}
}
