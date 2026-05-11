package tcp

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

type ProgressSyncServer struct {
	port        string
	Connections map[string]net.Conn
	mutex       sync.Mutex
}

func NewTCPServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		port:        port,
		Connections: make(map[string]net.Conn),
	}
}

func (s *ProgressSyncServer) Start() {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Fatalf("Failed to start TCP: %v", err)
	}
	defer lis.Close()
	log.Printf("📡 TCP Sync Server listening on :%s", s.port)

	for {
		conn, err := lis.Accept()
		if err == nil {
			go s.handleConnection(conn)
		}
	}
}

func (s *ProgressSyncServer) handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	s.mutex.Lock()
	s.Connections[addr] = conn
	s.mutex.Unlock()

	log.Printf("TCP client connected: %s", addr)
	buffer := make([]byte, 1024)
	for {
		if _, err := conn.Read(buffer); err != nil {
			break
		}
	}

	s.mutex.Lock()
	delete(s.Connections, addr)
	s.mutex.Unlock()
	conn.Close()
}

func (s *ProgressSyncServer) Broadcast(data interface{}) {
	msg, _ := json.Marshal(data)
	msg = append(msg, '\n')

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for addr, conn := range s.Connections {
		if _, err := conn.Write(msg); err != nil {
			conn.Close()
			delete(s.Connections, addr)
		}
	}
}