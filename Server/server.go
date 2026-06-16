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

func (s *Server) isUsernameTaken(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		if strings.EqualFold(c.username, username) {
			return true
		}
	}
	return false
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

	var username string
	for {
		fmt.Fprintln(writer, "[SERVER] Masukkan username Anda:")
		writer.Flush()

		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		nameInput = strings.TrimSpace(nameInput)

		if nameInput == "" {
			fmt.Fprintln(writer, "[SERVER] ERROR: Username tidak boleh kosong.")
			writer.Flush()
			continue
		}
		if s.isUsernameTaken(nameInput) {
			fmt.Fprintln(writer, "[SERVER] ERROR: Username '"+nameInput+"' sudah digunakan. Pilih username lain.")
			writer.Flush()
			continue
		}

		username = nameInput
		break
	}

	client := &Client{conn: conn, username: username, writer: writer}
	s.mu.Lock()
	s.clients[conn] = client
	s.mu.Unlock()

	fmt.Fprintln(writer, "[SERVER] Selamat datang, "+username+"!")
	writer.Flush()
	
	s.broadcast(fmt.Sprintf("[SERVER] %s bergabung ke chat\n", client.username), conn)
	

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
	s.broadcast(fmt.Sprintf("[SERVER] %s keluar dari chat\n", client.username), nil)
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
