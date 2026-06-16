package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tidak dapat terhubung: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Terhubung ke server!")

	serverReader := bufio.NewReader(conn)
	localReader := bufio.NewReader(os.Stdin)

	// Fase login
	for {
		line, _ := serverReader.ReadString('\n')
		msg := strings.TrimRight(line, "\r\n")
		fmt.Println(msg)

		if strings.Contains(msg, "Masukkan username") {
			fmt.Print(">> ")
			input, _ := localReader.ReadString('\n')
			fmt.Fprintln(conn, strings.TrimSpace(input))
		}
		if strings.Contains(msg, "Selamat datang") {
			// Baca satu baris lagi (baris info perintah)
			extra, _ := serverReader.ReadString('\n')
			fmt.Println(strings.TrimRight(extra, "\r\n"))
			break
		}
	}
	fmt.Println("Ketik pesan, /join <room>, /leave, atau /quit.")

	// Goroutine: terima pesan dari server
	go func() {
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("[Koneksi terputus]")
				os.Exit(0)
			}
			fmt.Print("\r" + strings.TrimRight(line, "\r\n") + "\n> ")
		}
	}()

	// Loop utama: kirim pesan ke server
	fmt.Print("> ")
	for {
		text, err := localReader.ReadString('\n')
		if err != nil {
			break
		}
		text = strings.TrimSpace(text)
		fmt.Fprintln(conn, text)
		if text == "/quit" {
			fmt.Println("Keluar dari chat. Sampai jumpa!")
			return
		}
		fmt.Print("> ")
	}
}