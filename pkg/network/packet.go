package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// PacketType defines the type of network packet
type PacketType uint8

const (
	// Connection packets
	PacketTypeHandshake PacketType = iota
	PacketTypeHandshakeAck
	PacketTypePing
	PacketTypePong
	PacketTypeDisconnect

	// Player packets
	PacketTypePlayerJoin
	PacketTypePlayerLeave
	PacketTypePlayerPosition
	PacketTypePlayerAction

	// World packets
	PacketTypeBlockChange
	PacketTypeChunkData
	PacketTypeWorldSync

	// Chat packets
	PacketTypeChatMessage

	// Discovery packets
	PacketTypeDiscoveryQuery
	PacketTypeDiscoveryResponse

	// Server info
	PacketTypeServerInfo
	PacketTypePlayerList
)

// PacketHeader contains metadata for all packets
type PacketHeader struct {
	Type      PacketType
	Timestamp uint64 // Unix nanoseconds
	Sequence  uint32 // Packet sequence number for ordering
	SenderID  string // Client/Server ID
}

// Packet represents a complete network message
type Packet struct {
	Header  PacketHeader
	Payload []byte
}

// PlayerPositionData contains player position update
type PlayerPositionData struct {
	PlayerID  string
	X, Y      float64
	VelocityX float64
	VelocityY float64
}

// PlayerActionData contains player actions (mining, placing, etc.)
type PlayerActionData struct {
	PlayerID   string
	ActionType string
	BlockX     int32
	BlockY     int32
	BlockType  uint8
}

// BlockChangeData represents a block change in the world
type BlockChangeData struct {
	X         int32
	Y         int32
	BlockType uint8
	PlayerID  string
}

// ChatMessageData represents a chat message
type ChatMessageData struct {
	PlayerID   string
	PlayerName string
	Message    string
	Timestamp  uint64
}

// PlayerInfo contains player metadata
type PlayerInfo struct {
	ID   string
	Name string
	X, Y float64
}

// ServerInfoData contains server information
type ServerInfoData struct {
	Name        string
	PlayerCount uint8
	MaxPlayers  uint8
	Address     string
	Port        uint16
}

// DiscoveryResponseData is sent by servers in response to queries
type DiscoveryResponseData struct {
	QueryID    string
	ServerInfo ServerInfoData
}

// NewPacket creates a new packet with the given type and sender
func NewPacket(packetType PacketType, senderID string) *Packet {
	return &Packet{
		Header: PacketHeader{
			Type:      packetType,
			Timestamp: uint64(time.Now().UnixNano()),
			Sequence:  0,
			SenderID:  senderID,
		},
	}
}

// Serialize converts the packet to bytes for transmission
func (p *Packet) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, p.Header.Type); err != nil {
		return nil, fmt.Errorf("failed to write packet type: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Header.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to write timestamp: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Header.Sequence); err != nil {
		return nil, fmt.Errorf("failed to write sequence: %w", err)
	}

	if err := writeString(buf, p.Header.SenderID); err != nil {
		return nil, fmt.Errorf("failed to write sender ID: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(p.Payload))); err != nil {
		return nil, fmt.Errorf("failed to write payload length: %w", err)
	}
	if len(p.Payload) > 0 {
		if _, err := buf.Write(p.Payload); err != nil {
			return nil, fmt.Errorf("failed to write payload: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// DeserializePacket parses a packet from bytes
func DeserializePacket(data []byte) (*Packet, error) {
	if len(data) < 14 {
		return nil, fmt.Errorf("packet too short")
	}

	buf := bytes.NewReader(data)
	p := &Packet{}

	if err := binary.Read(buf, binary.LittleEndian, &p.Header.Type); err != nil {
		return nil, fmt.Errorf("failed to read packet type: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Header.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Header.Sequence); err != nil {
		return nil, fmt.Errorf("failed to read sequence: %w", err)
	}

	senderID, err := readString(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read sender ID: %w", err)
	}
	p.Header.SenderID = senderID

	var payloadLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}

	if payloadLen > 0 {
		p.Payload = make([]byte, payloadLen)
		if _, err := buf.Read(p.Payload); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	return p, nil
}

// SerializePlayerPosition converts PlayerPositionData to bytes
func SerializePlayerPosition(data *PlayerPositionData) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := writeString(buf, data.PlayerID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.X); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.Y); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.VelocityX); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.VelocityY); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializePlayerPosition parses PlayerPositionData from bytes
func DeserializePlayerPosition(data []byte) (*PlayerPositionData, error) {
	buf := bytes.NewReader(data)
	p := &PlayerPositionData{}

	playerID, err := readString(buf)
	if err != nil {
		return nil, err
	}
	p.PlayerID = playerID

	if err := binary.Read(buf, binary.LittleEndian, &p.X); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Y); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.VelocityX); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.VelocityY); err != nil {
		return nil, err
	}

	return p, nil
}

// SerializeBlockChange converts BlockChangeData to bytes
func SerializeBlockChange(data *BlockChangeData) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, data.X); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.Y); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.BlockType); err != nil {
		return nil, err
	}
	if err := writeString(buf, data.PlayerID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeBlockChange parses BlockChangeData from bytes
func DeserializeBlockChange(data []byte) (*BlockChangeData, error) {
	buf := bytes.NewReader(data)
	b := &BlockChangeData{}

	if err := binary.Read(buf, binary.LittleEndian, &b.X); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &b.Y); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &b.BlockType); err != nil {
		return nil, err
	}

	playerID, err := readString(buf)
	if err != nil {
		return nil, err
	}
	b.PlayerID = playerID

	return b, nil
}

// SerializeChatMessage converts ChatMessageData to bytes
func SerializeChatMessage(data *ChatMessageData) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := writeString(buf, data.PlayerID); err != nil {
		return nil, err
	}
	if err := writeString(buf, data.PlayerName); err != nil {
		return nil, err
	}
	if err := writeString(buf, data.Message); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.Timestamp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeChatMessage parses ChatMessageData from bytes
func DeserializeChatMessage(data []byte) (*ChatMessageData, error) {
	buf := bytes.NewReader(data)
	c := &ChatMessageData{}

	playerID, err := readString(buf)
	if err != nil {
		return nil, err
	}
	c.PlayerID = playerID

	playerName, err := readString(buf)
	if err != nil {
		return nil, err
	}
	c.PlayerName = playerName

	message, err := readString(buf)
	if err != nil {
		return nil, err
	}
	c.Message = message

	if err := binary.Read(buf, binary.LittleEndian, &c.Timestamp); err != nil {
		return nil, err
	}

	return c, nil
}

// SerializeServerInfo converts ServerInfoData to bytes
func SerializeServerInfo(data *ServerInfoData) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := writeString(buf, data.Name); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.PlayerCount); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.MaxPlayers); err != nil {
		return nil, err
	}
	if err := writeString(buf, data.Address); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data.Port); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeServerInfo parses ServerInfoData from bytes
func DeserializeServerInfo(data []byte) (*ServerInfoData, error) {
	buf := bytes.NewReader(data)
	s := &ServerInfoData{}

	name, err := readString(buf)
	if err != nil {
		return nil, err
	}
	s.Name = name

	if err := binary.Read(buf, binary.LittleEndian, &s.PlayerCount); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &s.MaxPlayers); err != nil {
		return nil, err
	}

	address, err := readString(buf)
	if err != nil {
		return nil, err
	}
	s.Address = address

	if err := binary.Read(buf, binary.LittleEndian, &s.Port); err != nil {
		return nil, err
	}

	return s, nil
}

// Helper functions for string serialization
func writeString(buf *bytes.Buffer, s string) error {
	strBytes := []byte(s)
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(strBytes))); err != nil {
		return err
	}
	_, err := buf.Write(strBytes)
	return err
}

func readString(buf *bytes.Reader) (string, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	strBytes := make([]byte, length)
	if _, err := buf.Read(strBytes); err != nil {
		return "", err
	}
	return string(strBytes), nil
}

// GenerateID creates a unique identifier for clients/servers
func GenerateID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
