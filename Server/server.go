package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "sync"
)

type Client struct {
    conn   net.Conn
    writer *bufio.Writer
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
            fmt.Fprintln(c.writer, msg)
            c.writer.Flush()
        }
    }
}

func (s *Server) handleClient(conn net.Conn) {
    defer conn.Close()
    writer := bufio.NewWriter(conn)
    reader := bufio.NewReader(conn)

    s.mu.Lock()
    s.clients[conn] = &Client{conn: conn, writer: writer}
    s.mu.Unlock()

    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            break
        }
        s.broadcast(line, conn)
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