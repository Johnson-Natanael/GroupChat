package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
)

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!")
		os.Exit(1)
	} else {
		fmt.Println("Listening...")
	}
}