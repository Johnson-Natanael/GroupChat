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
	room     string // room saat ini, default "lobby"
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

func (s *Server) broadcastRoom(room, msg string, exclude net.Conn) {
	// Kumpulkan penerima di bawah lock, kirim di luar lock
	s.mu.Lock()
	targets := make([]*Client, 0)
	for _, c := range s.clients {
		if c.conn != exclude && c.room == room {
			targets = append(targets, c)
		}
	}
	s.mu.Unlock()

	for _, c := range targets {
		fmt.Fprint(c.writer, msg)
		c.writer.Flush()
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

	client := &Client{conn: conn, username: username, writer: writer, room: "lobby"}
	s.mu.Lock()
	s.clients[conn] = client
	s.mu.Unlock()

	fmt.Fprintln(writer, "[SERVER] Selamat datang, "+username+"! Anda berada di lobby.")
	fmt.Fprintln(writer, "[SERVER] Perintah: /join <room>  /leave  /quit")
	writer.Flush()

	s.broadcastRoom("lobby", fmt.Sprintf("[SERVER] %s bergabung ke lobby\n", username), conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case line == "/quit":
			goto disconnect

		case strings.HasPrefix(line, "/join "):
			newRoom := strings.TrimSpace(strings.TrimPrefix(line, "/join "))
			if newRoom == "" {
				fmt.Fprintln(writer, "[SERVER] Penggunaan: /join <nama_room>")
				writer.Flush()
				continue
			}
			oldRoom := client.room
			if oldRoom == newRoom {
				fmt.Fprintln(writer, "[SERVER] Anda sudah berada di room '"+newRoom+"'.")
				writer.Flush()
				continue
			}
			s.broadcastRoom(oldRoom, fmt.Sprintf("[SERVER] %s meninggalkan %s\n", username, oldRoom), conn)
			s.mu.Lock()
			client.room = newRoom
			s.mu.Unlock()
			fmt.Fprintln(writer, "[SERVER] Anda bergabung ke room '"+newRoom+"'.")
			writer.Flush()
			s.broadcastRoom(newRoom, fmt.Sprintf("[SERVER] %s bergabung ke %s\n", username, newRoom), conn)

		case line == "/leave":
			if client.room == "lobby" {
				fmt.Fprintln(writer, "[SERVER] Anda sudah berada di lobby.")
				writer.Flush()
				continue
			}
			oldRoom := client.room
			s.broadcastRoom(oldRoom, fmt.Sprintf("[SERVER] %s meninggalkan %s\n", username, oldRoom), conn)
			s.mu.Lock()
			client.room = "lobby"
			s.mu.Unlock()
			fmt.Fprintln(writer, "[SERVER] Anda kembali ke lobby.")
			writer.Flush()
			s.broadcastRoom("lobby", fmt.Sprintf("[SERVER] %s kembali ke lobby\n", username), conn)

		default:
			msg := fmt.Sprintf("[%s]: %s\n", username, line)
			s.broadcastRoom(client.room, msg, conn)
		}
	}

disconnect:
	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	s.broadcastRoom(client.room, fmt.Sprintf("[SERVER] %s keluar dari chat\n", username), nil)
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