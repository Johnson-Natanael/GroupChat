package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	//map untuk menyimpan koneksi dan username
	clients = make(map[net.Conn]string)
	//mengunci variabel mutex karena akan diakses oleh beberapa goroutine
	clientsMu sync.Mutex
)

// mengirimkan pesan ke seluruh client kecuali pengirim
func broadcast(message string, sender net.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if client != sender {
			fmt.Fprint(client, message)
		}
	}
}

// menangani setiap koneksi client secara paralel
func handleConnection(conn net.Conn) {
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
		fmt.Printf("Connection closed for %s\n", conn.RemoteAddr().String())
	}()

	fmt.Printf("New connection accepted: %s\n", conn.RemoteAddr().String())

	// megistrasi client baru
	var username string
	reader := bufio.NewReader(conn)

	nameInput, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	nameInput = strings.TrimSpace(nameInput)

	clientsMu.Lock()
	username = nameInput
	clients[conn] = username
	clientsMu.Unlock()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		// Format pesan dengan identitas username pengirim
		formattedMsg := fmt.Sprintf("[%s]: %s", username, message)
		fmt.Print(formattedMsg)
		broadcast(formattedMsg, conn)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen\n")
		os.Exit(1)
	} else {
		fmt.Println("Listening...")
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection\n")
			continue
		}
		// menggunakan goroutine untuk menangani setiap koneksi secara paralel
		go handleConnection(conn)
	}
}
