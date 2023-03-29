package tcp

import (
	"fmt"
	"time"

	"github.com/maurice2k/tcpserver"
	"github.com/rs/zerolog"
)

// Server is a TCP server for handling incoming device or retranslator connections.
type Server struct {
	srv     *tcpserver.Server
	logger  zerolog.Logger
	Handler *Handler
}

// NewServer creates a new TCP server.
func NewServer(l zerolog.Logger, address string) (server *Server, err error) {
	server = &Server{logger: l}
	server.srv, err = tcpserver.NewServer(address)
	server.srv.SetListenConfig(&tcpserver.ListenConfig{
		SocketReusePort:   true,
		SocketFastOpen:    false,
		SocketDeferAccept: false,
	})
	// server.srv.SetLoops(2000)
	// server.srv.SetWorkerpoolShards(8)
	// server.SetAllowThreadLocking(true)
	return
}

// RegisterHandler registers a handler for incoming device or retranslator connections.
func (s *Server) RegisterHandler(f tcpserver.RequestHandlerFunc) {
	s.srv.SetRequestHandler(f)
}

// SetProtocol sets the protocol for incoming device or retranslator connections.
func (s *Server) SetProtocol(p Protocol) {
	s.logger = s.logger.With().Str("proto", p.GetName()).Logger()
	s.Handler = NewHandler(s.logger, p, 10*time.Second) // ToDo make configurable
	s.RegisterHandler(s.Handler.Handle)
}

// ListenAndServe starts the TCP server.
func (s *Server) ListenAndServe() error {
	err := s.srv.Listen()

	if err != nil {
		return fmt.Errorf("error listening on interface: %w", err)
	}

	s.logger.Info().
		Str("local", s.srv.GetListenAddr().String()).
		Msg("server starts serving")
	err = s.srv.Serve()
	if err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}
