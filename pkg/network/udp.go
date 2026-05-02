package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	UDPBufferSize = 2048
	UDPTimeout    = 5 * time.Second
)

type UDPConn struct {
	conn        *net.UDPConn
	addr        *net.UDPAddr
	isServer    bool
	closed      bool
	closeMu     sync.RWMutex
	receiveChan chan *UDPMessage
	readMu      sync.Mutex
	sendMu      sync.Mutex
}

type UDPMessage struct {
	Data []byte
	Addr *net.UDPAddr
	Time time.Time
}

func NewUDPListener(port uint16) (*UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	if err := conn.SetReadBuffer(UDPBufferSize * 10); err != nil {
		log.Printf("Warning: failed to set UDP read buffer: %v", err)
	}

	uc := &UDPConn{
		conn:        conn,
		addr:        addr,
		isServer:    true,
		receiveChan: make(chan *UDPMessage, 100),
	}

	go uc.readLoop()
	return uc, nil
}

func NewUDPClient(serverAddr string, port uint16) (*UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverAddr, port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP: %w", err)
	}

	uc := &UDPConn{
		conn:        conn,
		addr:        addr,
		isServer:    false,
		receiveChan: make(chan *UDPMessage, 100),
	}

	go uc.readLoop()
	return uc, nil
}

func (uc *UDPConn) readLoop() {
	buffer := make([]byte, UDPBufferSize)
	for {
		uc.closeMu.RLock()
		if uc.closed {
			uc.closeMu.RUnlock()
			return
		}
		uc.closeMu.RUnlock()

		uc.readMu.Lock()
		uc.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, addr, err := uc.conn.ReadFromUDP(buffer)
		uc.readMu.Unlock()

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if !uc.IsClosed() {
				log.Printf("UDP read error: %v", err)
			}
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buffer[:n])
			msg := &UDPMessage{Data: data, Addr: addr, Time: time.Now()}
			select {
			case uc.receiveChan <- msg:
			default:
				log.Printf("UDP receive channel full, dropping packet")
			}
		}
	}
}

func (uc *UDPConn) Send(data []byte, addr *net.UDPAddr) error {
	uc.closeMu.RLock()
	if uc.closed {
		uc.closeMu.RUnlock()
		return fmt.Errorf("connection closed")
	}
	uc.closeMu.RUnlock()

	uc.sendMu.Lock()
	defer uc.sendMu.Unlock()

	if uc.isServer && addr != nil {
		_, err := uc.conn.WriteToUDP(data, addr)
		return err
	}
	_, err := uc.conn.Write(data)
	return err
}

func (uc *UDPConn) SendToServer(data []byte) error {
	if uc.isServer {
		return fmt.Errorf("cannot use SendToServer in server mode")
	}
	return uc.Send(data, nil)
}

func (uc *UDPConn) Receive() <-chan *UDPMessage {
	return uc.receiveChan
}

func (uc *UDPConn) LocalAddr() net.Addr {
	return uc.conn.LocalAddr()
}

func (uc *UDPConn) RemoteAddr() net.Addr {
	return uc.conn.RemoteAddr()
}

func (uc *UDPConn) Close() error {
	uc.closeMu.Lock()
	defer uc.closeMu.Unlock()
	if uc.closed {
		return nil
	}
	uc.closed = true
	close(uc.receiveChan)
	return uc.conn.Close()
}

func (uc *UDPConn) IsClosed() bool {
	uc.closeMu.RLock()
	defer uc.closeMu.RUnlock()
	return uc.closed
}

func (uc *UDPConn) SetReadDeadline(t time.Time) error {
	return uc.conn.SetReadDeadline(t)
}
