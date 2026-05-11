package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil { log.Fatal("Connection failed:", err) }
	defer conn.Close()

	fmt.Println("✅ Connected to TCP Sync Server! Listening for updates...")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Printf("\n🚀 [SYNC RECEIVED]: %s\n", scanner.Text())
	}
}