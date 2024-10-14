package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

const defaultListenAdds = ":5001"

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte
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
	slog.Info("server running on", s.ListenAddr)
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
	log.Fatal(server.Start())
}
