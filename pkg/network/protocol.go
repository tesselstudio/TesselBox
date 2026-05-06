package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"kaijuengine.com/matrix"
)

// MessageType represents different types of network messages
type MessageType uint8

const (
	MessageTypeHandshake MessageType = iota
	MessageTypeHandshakeResponse
	MessageTypePlayerJoin
	MessageTypePlayerLeave
	MessageTypePlayerMove
	MessageTypeBlockPlace
	MessageTypeBlockBreak
	MessageTypeChat
	MessageTypeWorldUpdate
	MessageTypePing
	MessageTypePong
	MessageTypeError
)

// Message represents a network message
type Message struct {
	Type      MessageType
	Data      []byte
	Timestamp uint64
	PlayerID  uint32
}

// HandshakeMessage is sent when a client first connects
type HandshakeMessage struct {
	Version    string
	PlayerName string
}

// HandshakeResponseMessage is sent by the server in response to handshake
type HandshakeResponseMessage struct {
	Success  bool
	PlayerID uint32
	Message  string
}

// PlayerJoinMessage is sent when a player joins the game
type PlayerJoinMessage struct {
	PlayerID uint32
	Name     string
	Position matrix.Vec3
}

// PlayerLeaveMessage is sent when a player leaves the game
type PlayerLeaveMessage struct {
	PlayerID uint32
}

// PlayerMoveMessage is sent when a player moves
type PlayerMoveMessage struct {
	PlayerID uint32
	Position matrix.Vec3
	Rotation matrix.Vec3
	Velocity matrix.Vec3
}

// BlockPlaceMessage is sent when a player places a block
type BlockPlaceMessage struct {
	PlayerID uint32
	BlockType uint8
	Position matrix.Vec3
	Rotation int
}

// BlockBreakMessage is sent when a player breaks a block
type BlockBreakMessage struct {
	PlayerID uint32
	Position matrix.Vec3
}

// ChatMessage is sent for chat communication
type ChatMessage struct {
	PlayerID uint32
	Message  string
}

// WorldUpdateMessage contains world state changes
type WorldUpdateMessage struct {
	BlockChanges []BlockChange
	PlayerStates  []PlayerState
}

// BlockChange represents a single block change
type BlockChange struct {
	Type     uint8 // 0 = break, 1 = place
	BlockType uint8
	Position matrix.Vec3
	Rotation int
}

// PlayerState represents a player's current state
type PlayerState struct {
	PlayerID  uint32
	Position  matrix.Vec3
	Rotation  matrix.Vec3
	Velocity  matrix.Vec3
	Health    uint16
}

// PingMessage is used for connection health checking
type PingMessage struct {
	Timestamp uint64
}

// PongMessage is the response to a ping
type PongMessage struct {
	Timestamp uint64
}

// ErrorMessage is sent when an error occurs
type ErrorMessage struct {
	Code    uint16
	Message string
}

// Protocol handles message encoding and decoding
type Protocol struct {
	buffer []byte
}

// NewProtocol creates a new protocol instance
func NewProtocol() *Protocol {
	return &Protocol{
		buffer: make([]byte, 0, 1024),
	}
}

// EncodeMessage encodes a message into bytes
func (p *Protocol) EncodeMessage(msg *Message) ([]byte, error) {
	p.buffer = p.buffer[:0] // Reset buffer
	
	// Write header
	p.buffer = append(p.buffer, byte(msg.Type))
	p.buffer = binary.BigEndian.AppendUint64(p.buffer, msg.Timestamp)
	p.buffer = binary.BigEndian.AppendUint32(p.buffer, msg.PlayerID)
	
	// Write data
	p.buffer = append(p.buffer, msg.Data...)
	
	// Write length prefix
	length := len(p.buffer)
	result := make([]byte, length+4)
	binary.BigEndian.PutUint32(result[:4], uint32(length))
	copy(result[4:], p.buffer)
	
	return result, nil
}

// DecodeMessage decodes a message from bytes
func (p *Protocol) DecodeMessage(data []byte) (*Message, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("message too short")
	}
	
	length := binary.BigEndian.Uint32(data[:4])
	if len(data) < int(length)+4 {
		return nil, fmt.Errorf("incomplete message")
	}
	
	data = data[4:] // Remove length prefix
	
	if len(data) < 13 { // Minimum header size
		return nil, fmt.Errorf("message header too short")
	}
	
	msg := &Message{
		Type:      MessageType(data[0]),
		Timestamp: binary.BigEndian.Uint64(data[1:9]),
		PlayerID:  binary.BigEndian.Uint32(data[9:13]),
		Data:      make([]byte, len(data)-13),
	}
	
	copy(msg.Data, data[13:])
	
	return msg, nil
}

// EncodeHandshake encodes a handshake message
func (p *Protocol) EncodeHandshake(handshake *HandshakeMessage) []byte {
	data := make([]byte, 0, 64)
	
	// Encode version
	versionBytes := []byte(handshake.Version)
	data = append(data, byte(len(versionBytes)))
	data = append(data, versionBytes...)
	
	// Encode player name
	nameBytes := []byte(handshake.PlayerName)
	data = append(data, byte(len(nameBytes)))
	data = append(data, nameBytes...)
	
	return data
}

// DecodeHandshake decodes a handshake message
func (p *Protocol) DecodeHandshake(data []byte) (*HandshakeMessage, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("handshake data too short")
	}
	
	versionLen := int(data[0])
	if len(data) < 1+versionLen+1 {
		return nil, fmt.Errorf("invalid handshake format")
	}
	
	version := string(data[1 : 1+versionLen])
	nameLen := int(data[1+versionLen])
	if len(data) < 1+versionLen+1+nameLen {
		return nil, fmt.Errorf("invalid handshake format")
	}
	
	name := string(data[2+versionLen : 2+versionLen+nameLen])
	
	return &HandshakeMessage{
		Version:    version,
		PlayerName: name,
	}, nil
}

// EncodeHandshakeResponse encodes a handshake response
func (p *Protocol) EncodeHandshakeResponse(resp *HandshakeResponseMessage) []byte {
	data := make([]byte, 0, 32)
	
	// Success flag
	if resp.Success {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	
	// Player ID
	data = binary.BigEndian.AppendUint32(data, resp.PlayerID)
	
	// Message
	msgBytes := []byte(resp.Message)
	data = append(data, byte(len(msgBytes)))
	data = append(data, msgBytes...)
	
	return data
}

// DecodeHandshakeResponse decodes a handshake response
func (p *Protocol) DecodeHandshakeResponse(data []byte) (*HandshakeResponseMessage, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("handshake response too short")
	}
	
	success := data[0] == 1
	playerID := binary.BigEndian.Uint32(data[1:5])
	msgLen := int(data[5])
	
	if len(data) < 6+msgLen {
		return nil, fmt.Errorf("invalid handshake response format")
	}
	
	message := string(data[6 : 6+msgLen])
	
	return &HandshakeResponseMessage{
		Success:  success,
		PlayerID: playerID,
		Message:  message,
	}, nil
}

// EncodePlayerMove encodes a player move message
func (p *Protocol) EncodePlayerMove(move *PlayerMoveMessage) []byte {
	data := make([]byte, 0, 36)
	
	// Player ID
	data = binary.BigEndian.AppendUint32(data, move.PlayerID)
	
	// Position
	data = binary.BigEndian.AppendUint32(data, uint32(move.Position.X()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Position.Y()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Position.Z()*1000))
	
	// Rotation
	data = binary.BigEndian.AppendUint32(data, uint32(move.Rotation.X()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Rotation.Y()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Rotation.Z()*1000))
	
	// Velocity
	data = binary.BigEndian.AppendUint32(data, uint32(move.Velocity.X()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Velocity.Y()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(move.Velocity.Z()*1000))
	
	return data
}

// DecodePlayerMove decodes a player move message
func (p *Protocol) DecodePlayerMove(data []byte) (*PlayerMoveMessage, error) {
	if len(data) < 36 {
		return nil, fmt.Errorf("player move data too short")
	}
	
	playerID := binary.BigEndian.Uint32(data[0:4])
	
	posX := matrix.Float(binary.BigEndian.Uint32(data[4:8])) / 1000
	posY := matrix.Float(binary.BigEndian.Uint32(data[8:12])) / 1000
	posZ := matrix.Float(binary.BigEndian.Uint32(data[12:16])) / 1000
	
	rotX := matrix.Float(binary.BigEndian.Uint32(data[16:20])) / 1000
	rotY := matrix.Float(binary.BigEndian.Uint32(data[20:24])) / 1000
	rotZ := matrix.Float(binary.BigEndian.Uint32(data[24:28])) / 1000
	
	velX := matrix.Float(binary.BigEndian.Uint32(data[28:32])) / 1000
	velY := matrix.Float(binary.BigEndian.Uint32(data[32:36])) / 1000
	velZ := matrix.Float(binary.BigEndian.Uint32(data[36:40])) / 1000
	
	return &PlayerMoveMessage{
		PlayerID: playerID,
		Position: matrix.NewVec3(posX, posY, posZ),
		Rotation: matrix.NewVec3(rotX, rotY, rotZ),
		Velocity: matrix.NewVec3(velX, velY, velZ),
	}, nil
}

// EncodeBlockPlace encodes a block place message
func (p *Protocol) EncodeBlockPlace(place *BlockPlaceMessage) []byte {
	data := make([]byte, 0, 20)
	
	// Player ID
	data = binary.BigEndian.AppendUint32(data, place.PlayerID)
	
	// Block type
	data = append(data, place.BlockType)
	
	// Position
	data = binary.BigEndian.AppendUint32(data, uint32(place.Position.X()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(place.Position.Y()*1000))
	data = binary.BigEndian.AppendUint32(data, uint32(place.Position.Z()*1000))
	
	// Rotation
	data = binary.BigEndian.AppendUint32(data, uint32(place.Rotation))
	
	return data
}

// DecodeBlockPlace decodes a block place message
func (p *Protocol) DecodeBlockPlace(data []byte) (*BlockPlaceMessage, error) {
	if len(data) < 17 {
		return nil, fmt.Errorf("block place data too short")
	}
	
	playerID := binary.BigEndian.Uint32(data[0:4])
	blockType := data[4]
	
	posX := matrix.Float(binary.BigEndian.Uint32(data[5:9])) / 1000
	posY := matrix.Float(binary.BigEndian.Uint32(data[9:13])) / 1000
	posZ := matrix.Float(binary.BigEndian.Uint32(data[13:17])) / 1000
	
	rotation := int(binary.BigEndian.Uint32(data[17:21]))
	
	return &BlockPlaceMessage{
		PlayerID:  playerID,
		BlockType: blockType,
		Position:  matrix.NewVec3(posX, posY, posZ),
		Rotation:  rotation,
	}, nil
}

// MessageReader handles reading messages from a connection
type MessageReader struct {
	reader io.Reader
	buffer []byte
	protocol *Protocol
}

// NewMessageReader creates a new message reader
func NewMessageReader(reader io.Reader) *MessageReader {
	return &MessageReader{
		reader:   reader,
		buffer:   make([]byte, 0, 4096),
		protocol:  NewProtocol(),
	}
}

// ReadMessage reads a complete message from the connection
func (mr *MessageReader) ReadMessage() (*Message, error) {
	// Read message length
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(mr.reader, lengthBytes)
	if err != nil {
		return nil, err
	}
	
	length := binary.BigEndian.Uint32(lengthBytes)
	if length > 4096 { // Max message size (4KB)
		return nil, fmt.Errorf("message too large: %d bytes (max %d)", length, 4096)
	}
	
	// Read message data
	messageBytes := make([]byte, length)
	_, err = io.ReadFull(mr.reader, messageBytes)
	if err != nil {
		return nil, err
	}
	
	// Decode message
	return mr.protocol.DecodeMessage(messageBytes)
}

// MessageWriter handles writing messages to a connection
type MessageWriter struct {
	writer   io.Writer
	protocol *Protocol
}

// NewMessageWriter creates a new message writer
func NewMessageWriter(writer io.Writer) *MessageWriter {
	return &MessageWriter{
		writer:   writer,
		protocol: NewProtocol(),
	}
}

// WriteMessage writes a message to the connection
func (mw *MessageWriter) WriteMessage(msg *Message) error {
	data, err := mw.protocol.EncodeMessage(msg)
	if err != nil {
		return err
	}
	
	_, err = mw.writer.Write(data)
	return err
}
