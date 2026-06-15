package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", ":9090")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Tidak dapat terhubung: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close()
    fmt.Println("Terhubung ke server!")

    // Goroutine: terima pesan dari server
    go func() {
        reader := bufio.NewReader(conn)
        for {
            line, err := reader.ReadString('\n')
            if err != nil {
                fmt.Println("[Koneksi terputus]")
                os.Exit(0)
            }
            fmt.Print(line)
        }
    }()

    // Loop utama: kirim pesan ke server
    reader := bufio.NewReader(os.Stdin)
    for {
        text, err := reader.ReadString('\n')
        if err != nil {
            break
        }
        fmt.Fprintln(conn, text)
    }
}