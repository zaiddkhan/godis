package main

import (
	"context"
	"fmt"
	client2 "godis/client"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"
)

const defaultListenAdds = ":5001"

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte

	mu sync.Mutex
	kv KV
}

type Config struct {
	ListenAddr string
	ln         net.Listener
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAdds
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
	}
}

func (s *Server) handleRawMessage(rawMsg []byte) error {
	fmt.Println(string(rawMsg))
	cmd, err := parseCommand(string(rawMsg))
	if err != nil {
		return err
	}
	switch v := cmd.(type) {
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	}
	return nil
}
func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil {
				log.Println("handleRawMessage:", err)
			}
			fmt.Println(rawMsg)
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true

		}
	}
}
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.loop()
	//slog.Info("server running on", s.ListenAddr)
	return s.acceptLoop()

}

func (s *Server) acceptLoop() error {

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error:", err)
			continue
		}
		go s.handleConn(conn)

	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected:", "remote addr", conn.RemoteAddr())

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error:", err)
	}
}

func main() {

	server := NewServer(Config{})

	go func() {
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)

	client := client2.NewClient("localhost:5001")
	count := 10
	for i := 0; i < count; i++ {
		if err := client.Set(context.Background(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			slog.Error("set client error:", err)
		}
	}
	time.Sleep(time.Second)
	fmt.Println(server.kv.data)

	select {}

}
