package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// menangani pembacaan pesan dari server secara paralel
func readFromServer(conn net.Conn) {
	connReader := bufio.NewReader(conn)
	for {
		message, err := connReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nConnection lost from server\n")
			os.Exit(1)
		}
		fmt.Print(message)
	}
}

func main() {
	//dial ke server
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server\n")
		os.Exit(1)
	} else {
		fmt.Println("Connected to serve")
	}
	defer conn.Close()

	//jalankan pembacaan pesan dari server secara paralel menggunakan goroutine
	go readFromServer(conn)

	localReader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, err := localReader.ReadString('\n')
	if err != nil {
		os.Exit(1)
	}

	fmt.Fprint(conn, username)

	for {
		message, err := localReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the message\n")
			break
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send the message\n")
			break
		}
	}
}
