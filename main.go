package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quit       chan struct{}
	msgch      chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quit:       make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.accpetLoop()

	<-s.quit
	close(s.msgch)
	return nil
}

func (s *Server) accpetLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Printf("accept Error: %s\n", err)
			continue
		}

		fmt.Println("new connection to the server:", conn.RemoteAddr())
		conn.Write([]byte("Welcome to my tcp server, Hi\n"))

		go s.readLoop(conn) // without go no concurrent connection will be possible as the accept loop will wait for the read to end to start another one

	}
}

func (s *Server) readLoop(conn net.Conn) {
	buf := make([]byte, 2048)
	defer conn.Close()

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("read Error: %s\n", err)
			continue

		}
		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}
	}
}

func main() {
	server := NewServer(":42069")

	go func() {
		for msg := range server.msgch {
			fmt.Printf("recieved message from connection (%s):%s", msg.from, msg.payload)
		}
	}()

	log.Fatal(server.Start())

}
