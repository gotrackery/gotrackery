package server

import (
	"fmt"
	"time"

	"github.com/maurice2k/tcpserver"
	"github.com/rs/zerolog"
)

type Option func(*Server)

// Server is a TCP server for handling incoming device or retranslator connections.
type Server struct {
	Handler *Handler
	srv     *tcpserver.Server
	logger  zerolog.Logger
	timeout time.Duration
}

// NewServer creates a new TCP server.
func NewServer(l zerolog.Logger, address string, opts ...Option) (server *Server, err error) {
	server = &Server{logger: l}
	server.srv, err = tcpserver.NewServer(address)
	server.srv.SetListenConfig(&tcpserver.ListenConfig{
		SocketReusePort: true,
	})
	server.Option(opts...)
	return
}

// Option sets the options specified.
func (s *Server) Option(opts ...Option) {
	for _, opt := range opts {
		opt(s)
	}
}

func WithTimeout(to time.Duration) Option {
	return func(s *Server) {
		s.timeout = to
	}
}

func WithSocketReusePort(reuse bool) Option {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketReusePort = reuse
		s.srv.SetListenConfig(c)
	}
}

func WithSocketFastOpen(fopen bool) Option {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketFastOpen = fopen
		s.srv.SetListenConfig(c)
	}
}

func WithSocketDeferAccept(daccept bool) Option {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketDeferAccept = daccept
		s.srv.SetListenConfig(c)
	}
}

func WithLoops(loops int) Option {
	return func(s *Server) {
		s.srv.SetLoops(loops)
	}
}

func WithWorkerpoolShards(shards int) Option {
	return func(s *Server) {
		s.srv.SetWorkerpoolShards(shards)
	}
}

func WithAllowThreadLocking(locking bool) Option {
	return func(s *Server) {
		s.srv.SetAllowThreadLocking(locking)
	}
}

// RegisterHandler registers a handler for incoming device or retranslator connections.
func (s *Server) RegisterHandler(f tcpserver.RequestHandlerFunc) {
	s.srv.SetRequestHandler(f)
}

// SetProtocol sets the protocol for incoming device or retranslator connections.
func (s *Server) SetProtocol(p Protocol) {
	s.logger = s.logger.With().Str("proto", p.GetName()).Logger()
	s.Handler = NewHandler(s.logger, p, s.timeout)
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
