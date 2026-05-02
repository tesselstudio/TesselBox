package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	TCPTimeout       = 10 * time.Second
	TCPReadTimeout   = 30 * time.Second
	TCPWriteTimeout  = 5 * time.Second
)

type TCPConn struct {
	conn        net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	closed      bool
	closeMu     sync.RWMutex
	receiveChan chan []byte
	sendChan    chan []byte
	onDisconnect func()
}

type TCPListener struct {
	listener   net.Listener
	closed     bool
	closeMu    sync.RWMutex
	onConnect  func(*TCPConn)
}

func NewTCPListener(port uint16) (*TCPListener, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on TCP: %w", err)
	}

	tl := &TCPListener{listener: listener}
	go tl.acceptLoop()
	return tl, nil
}

func (tl *TCPListener) acceptLoop() {
	for {
		tl.closeMu.RLock()
		if tl.closed {
			tl.closeMu.RUnlock()
			return
		}
		tl.closeMu.RUnlock()

		tcpConn, err := tl.listener.Accept()
		if err != nil {
			tl.closeMu.RLock()
			closed := tl.closed
			tl.closeMu.RUnlock()
			if closed {
				return
			}
			log.Printf("TCP accept error: %v", err)
			continue
		}

		tcpConn.SetDeadline(time.Now().Add(TCPTimeout))

		conn := &TCPConn{
			conn:        tcpConn,
			reader:      bufio.NewReader(tcpConn),
			writer:      bufio.NewWriter(tcpConn),
			receiveChan: make(chan []byte, 100),
			sendChan:    make(chan []byte, 100),
		}

		go conn.readLoop()
		go conn.writeLoop()

		if tl.onConnect != nil {
			tl.onConnect(conn)
		}
	}
}

func (tl *TCPListener) OnConnect(callback func(*TCPConn)) {
	tl.onConnect = callback
}

func (tl *TCPListener) Close() error {
	tl.closeMu.Lock()
	defer tl.closeMu.Unlock()
	if tl.closed {
		return nil
	}
	tl.closed = true
	return tl.listener.Close()
}

func (tl *TCPListener) Addr() net.Addr {
	return tl.listener.Addr()
}

func NewTCPClient(serverAddr string, port uint16) (*TCPConn, error) {
	addr := fmt.Sprintf("%s:%d", serverAddr, port)
	conn, err := net.DialTimeout("tcp", addr, TCPTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	tcpConn := &TCPConn{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriter(conn),
		receiveChan: make(chan []byte, 100),
		sendChan:    make(chan []byte, 100),
	}

	go tcpConn.readLoop()
	go tcpConn.writeLoop()

	return tcpConn, nil
}

func (tc *TCPConn) readLoop() {
	defer close(tc.receiveChan)
	for {
		tc.closeMu.RLock()
		if tc.closed {
			tc.closeMu.RUnlock()
			return
		}
		tc.closeMu.RUnlock()

		tc.conn.SetReadDeadline(time.Now().Add(TCPReadTimeout))

		var length uint32
		if err := binary.Read(tc.reader, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if !tc.IsClosed() {
				log.Printf("TCP read error: %v", err)
			}
			break
		}

		if length > 0 && length < 1024*1024 {
			data := make([]byte, length)
			if _, err := io.ReadFull(tc.reader, data); err != nil {
				if !tc.IsClosed() {
					log.Printf("TCP read data error: %v", err)
				}
				break
			}
			select {
			case tc.receiveChan <- data:
			default:
				log.Printf("TCP receive channel full, dropping message")
			}
		}
	}
	if tc.onDisconnect != nil {
		tc.onDisconnect()
	}
}

func (tc *TCPConn) writeLoop() {
	for data := range tc.sendChan {
		tc.closeMu.RLock()
		if tc.closed {
			tc.closeMu.RUnlock()
			return
		}
		tc.closeMu.RUnlock()

		tc.conn.SetWriteDeadline(time.Now().Add(TCPWriteTimeout))

		if err := binary.Write(tc.writer, binary.LittleEndian, uint32(len(data))); err != nil {
			log.Printf("TCP write length error: %v", err)
			return
		}
		if _, err := tc.writer.Write(data); err != nil {
			log.Printf("TCP write data error: %v", err)
			return
		}
		if err := tc.writer.Flush(); err != nil {
			log.Printf("TCP flush error: %v", err)
			return
		}
	}
}

func (tc *TCPConn) Send(data []byte) error {
	tc.closeMu.RLock()
	if tc.closed {
		tc.closeMu.RUnlock()
		return fmt.Errorf("connection closed")
	}
	tc.closeMu.RUnlock()

	select {
	case tc.sendChan <- data:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}

func (tc *TCPConn) SendPacket(packet *Packet) error {
	data, err := packet.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %w", err)
	}
	return tc.Send(data)
}

func (tc *TCPConn) Receive() <-chan []byte {
	return tc.receiveChan
}

func (tc *TCPConn) Close() error {
	tc.closeMu.Lock()
	defer tc.closeMu.Unlock()
	if tc.closed {
		return nil
	}
	tc.closed = true
	close(tc.sendChan)
	return tc.conn.Close()
}

func (tc *TCPConn) IsClosed() bool {
	tc.closeMu.RLock()
	defer tc.closeMu.RUnlock()
	return tc.closed
}

func (tc *TCPConn) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

func (tc *TCPConn) LocalAddr() net.Addr {
	return tc.conn.LocalAddr()
}

func (tc *TCPConn) OnDisconnect(callback func()) {
	tc.onDisconnect = callback
}
