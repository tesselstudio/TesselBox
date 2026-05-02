package server

import (
	"fmt"
	"log"
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Port      int
	MaxPlayers int
	WorldName string
}

// DefaultConfig returns default server configuration
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Port:       25565,
		MaxPlayers: 100,
		WorldName:  "world",
	}
}

// Server represents the game server
type Server struct {
	config *ServerConfig
	listener net.Listener
	running bool
}

// NewServer creates a new server instance
func NewServer(config *ServerConfig) *Server {
	return &Server{
		config:  config,
		running: false,
	}
}

// Run starts the server
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener
	s.running = true

	log.Printf("Server started on port %d", s.config.Port)
	log.Printf("Max players: %d", s.config.MaxPlayers)
	log.Printf("World: %s", s.config.WorldName)

	// Accept connections
	for s.running {
		conn, err := listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Accept error: %v", err)
			}
			continue
		}
		go s.handleConnection(conn)
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	return s.running
}

// GetPlayerCount returns current player count
func (s *Server) GetPlayerCount() int {
	return 0 // Placeholder
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())
	// Handle client connection
}

// TUIModel represents the TUI model for the server
type TUIModel struct {
	Server  *Server
	Choices []string
	Cursor  int
	Started bool
}

// Init initializes the TUI model
func (m TUIModel) Init() tea.Cmd {
	return nil
}

// Update handles TUI updates
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter":
			if m.Cursor == 0 && !m.Started {
				m.Started = true
				go m.Server.Run()
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			} else if m.Cursor == 1 {
				return m, tea.Quit
			}
		}
	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

type tickMsg time.Time

// View renders the TUI
func (m TUIModel) View() string {
	s := "TesselBox Server\n\n"

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		status := ""
		if i == 0 && m.Started {
			status = " [RUNNING]"
		}

		s += fmt.Sprintf("%s %s%s\n", cursor, choice, status)
	}

	s += "\nPress q to quit\n"

	if m.Started {
		s += fmt.Sprintf("\nPlayers: %d/%d\n", m.Server.GetPlayerCount(), m.Server.config.MaxPlayers)
	}

	return s
}
