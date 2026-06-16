package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	conn     net.Conn
	writer   *bufio.Writer
	username string
}

type Server struct {
	mu      sync.Mutex
	clients map[net.Conn]*Client
}

func NewServer() *Server {
	return &Server{clients: make(map[net.Conn]*Client)}
}

func (s *Server) broadcast(msg string, exclude net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		if c.conn != exclude {
			fmt.Fprint(c.writer, msg)
			c.writer.Flush()
		}
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	for {
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		nameInput = strings.TrimSpace(nameInput)

		if nameInput == "" {
			fmt.Fprint(writer, "Username tidak boleh kosong")
			writer.Flush()
			continue
		}

		s.mu.Lock()
		taken := false

		for _, c := range s.clients {
			if c.username == nameInput {
				taken = true
				break
			}
		}

		if !taken {
			s.clients[conn] = &Client{conn: conn, writer: writer, username: nameInput}
			s.mu.Unlock()
			writer.Flush()
			break
		}
		s.mu.Unlock()
		fmt.Fprintln(writer, "Username sudah digunakan")
		writer.Flush()
	}

	client := s.clients[conn]

	s.broadcast(fmt.Sprintf("%s bergabung ke chat\n", client.username), conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		formattedMsg := fmt.Sprintf("[%s]: %s", client.username, line)
		s.broadcast(formattedMsg, conn)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	s.broadcast(fmt.Sprintf("%s keluar dari chat\n", client.username), nil)
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Gagal mendengarkan: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("[SERVER] Berjalan di :9090")
	fmt.Println("[SERVER] Menunggu koneksi...")

	srv := NewServer()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go srv.handleClient(conn)
	}
}
