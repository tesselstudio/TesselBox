package network

import (
	"net"
	"sync"
	"time"
)

type ClientSession struct {
	ID       string
	Name     string
	PlayerID string
	UDPAddr  *net.UDPAddr
	TCPConn  *TCPConn
	PositionX, PositionY, VelocityX, VelocityY float64
	IsConnected    bool
	LastSeen       time.Time
	JoinTime       time.Time
	reliableSendMu sync.Mutex
	PlayerData     map[string]interface{}
	dataMu         sync.RWMutex
}

func NewClientSession(id, name, playerID string) *ClientSession {
	return &ClientSession{
		ID:          id,
		Name:        name,
		PlayerID:    playerID,
		IsConnected: true,
		LastSeen:    time.Now(),
		JoinTime:    time.Now(),
		PlayerData:  make(map[string]interface{}),
	}
}

func (cs *ClientSession) UpdatePosition(x, y, vx, vy float64) {
	cs.PositionX = x
	cs.PositionY = y
	cs.VelocityX = vx
	cs.VelocityY = vy
	cs.LastSeen = time.Now()
}

func (cs *ClientSession) GetPosition() (x, y, vx, vy float64) {
	return cs.PositionX, cs.PositionY, cs.VelocityX, cs.VelocityY
}

func (cs *ClientSession) SetPlayerData(key string, value interface{}) {
	cs.dataMu.Lock()
	defer cs.dataMu.Unlock()
	cs.PlayerData[key] = value
}

func (cs *ClientSession) GetPlayerData(key string) (interface{}, bool) {
	cs.dataMu.RLock()
	defer cs.dataMu.RUnlock()
	val, ok := cs.PlayerData[key]
	return val, ok
}

func (cs *ClientSession) IsTimedOut(timeout time.Duration) bool {
	return time.Since(cs.LastSeen) > timeout
}

func (cs *ClientSession) Disconnect() {
	cs.IsConnected = false
	if cs.TCPConn != nil {
		cs.TCPConn.Close()
	}
}

func (cs *ClientSession) SendReliable(packet *Packet) error {
	if cs.TCPConn == nil || cs.TCPConn.IsClosed() {
		return nil
	}
	return cs.TCPConn.SendPacket(packet)
}

func (cs *ClientSession) ToPlayerInfo() *PlayerInfo {
	return &PlayerInfo{
		ID:   cs.PlayerID,
		Name: cs.Name,
		X:    cs.PositionX,
		Y:    cs.PositionY,
	}
}
