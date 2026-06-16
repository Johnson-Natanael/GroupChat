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
	username string
	writer   *bufio.Writer
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
			fmt.Fprintln(c.writer, msg)
			c.writer.Flush()
		}
	}
}

// fungsi untuk menangani tiap client yang terhubung
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Fase login: minta username
	var username string
	for {
		fmt.Fprintln(writer, "[SERVER] Masukkan username Anda:")
		writer.Flush()

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		name := strings.TrimSpace(line)

		if name == "" {
			fmt.Fprintln(writer, "[SERVER] ERROR: Username tidak boleh kosong.")
			writer.Flush()
			continue
		}
		if s.isUsernameTaken(name) {
			fmt.Fprintln(writer, "[SERVER] ERROR: Username '"+name+"' sudah digunakan. Pilih username lain.")
			writer.Flush()
			continue
		}

		username = name
		break
	}

	// ketika username valid dan client berhasil login, buat struct Client dan daftarkan ke server
	client := &Client{conn: conn, username: username, writer: writer}
	s.mu.Lock()
	s.clients[conn] = client
	s.mu.Unlock()
	// Setelah register client — notifikasi JOIN
	s.broadcast("[SERVER] "+username+" telah bergabung.", conn)

	fmt.Fprintln(writer, "[SERVER] Selamat datang, "+username+"!")
	writer.Flush()

	// Loop pesan untuk dibroadcast ke semua client lain
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msg := "[" + username + "] " + strings.TrimSpace(line)
		fmt.Println(msg)
		s.broadcast(msg, conn)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Gagal mendengarkan: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("[SERVER] Berjalan di :9090")

	srv := NewServer()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go srv.handleClient(conn)
	}
}
