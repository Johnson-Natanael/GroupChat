package main

import (
	"fmt"
	"sync"
)

type room struct {
	mu      sync.Mutex       // Mutex biar gk data race pas banyak user masuk/keluar
	name    string           // Nama buat room
	clients map[*Client]bool //pointer client yang lagi di room
}

// global variable buat cek room
var (
	roomsMu sync.Mutex // Mutex biar gk data race pas bikin room
	rooms   = make(map[string]*room)
)

// function buat room
func NewRoom(name string) *room {
	return &room{
		name:    name,
		clients: make(map[*Client]bool), // Alokasi memori untuk map kosong agar tidak terjadi nil-pointer panics.
	}
}

// function room manager
func GetorCreateRoom(name string) *room {
	roomsMu.Lock()
	defer roomsMu.Unlock() // unlock setelah fungsi selesai dieksekusi

	// cek jika ada nama room yg sama
	if room, exists := rooms[name]; exists {
		return room // kembalikan pointer room yg sudah ada
	}

	// buat objek room baru
	room := NewRoom(name)
	rooms[name] = room //masukan nama room baru ke map
	return room
}

// function buat ngecek semua room yg ada
func GetActiveRooms() string {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// kalo blm ada room sama sekali
	if len(rooms) == 0 {
		return "[SERVER] Belum ada room yang aktif \n"
	}

	result := "[SERVER] Daftar Room Yang Aktif:\n"

	// loop untuk mengambil semua room yg ada
	for roomName, room := range rooms {
		room.mu.Lock()
		totalClients := len(room.clients) // Hitung total orang di dalem room
		room.mu.Unlock()

		// output
		result += fmt.Sprintf(" - #%s (%d orang)\n", roomName, totalClients)
	}
	return result
}

// mekanik join room
func (r *room) Join(c *Client) {
	r.mu.Lock()
	r.clients[c] = true // tambah client ke map room
	r.mu.Unlock()

	c.currentRoom = r // set room aktif si client ke room ini

	r.Broadcast(fmt.Sprintf("[SERVER] %s Bergabung Ke Room #&s\n", c.username, r.name), nil)

}

// mekanik leave room
func (r *room) Leave(c *Client) {
	r.mu.Lock()

	// cek kalo user gk ada di room
	if _, exists := r.clients[c]; exists {
		delete(r.clients, c) // delete user dari room
	}
	r.mu.Unlock()

	c.currentRoom = nil // reset room pada objek client, kembali ke lobby

	// Notifikasi kalo ada yg leave
	r.Broadcast(fmt.Sprintf("[SERVER] %s Keluar Dari Room #%s\n", c.username, r.name), nil)
}

// function mengirimkan pesan ke semua anggota yg ada di room, exclude client yg gk masuk room
func (r *room) Broadcast(msg string, exclude *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// loop untuk setiap client yg ada di room
	for c := range r.clients {
		// Jika client saat ini bukan client yg kena exclude
		if c != exclude {
			fmt.Fprint(c.writer, msg) // Tulis pesan ke buffer client
			c.writer.Flush()          // Paksa buffer buat ngirim pesan pake client
		}
	}
}
