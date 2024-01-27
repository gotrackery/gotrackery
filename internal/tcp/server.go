package tcp

import (
	"context"
	"fmt"
	"time"

	"github.com/gotrackery/gotrackery/internal/event"
	"github.com/maurice2k/tcpserver"
	"github.com/rs/zerolog"
)

type ServerOption func(*Server)

// Server is a TCP server for handling incoming device or retranslator connections.
type Server struct {
	Handler *Handler
	srv     *tcpserver.Server
	logger  zerolog.Logger
	timeout time.Duration
}

// NewServer creates a new TCP server.
func NewServer(l zerolog.Logger, address string, opts ...ServerOption) (server *Server, err error) {
	server = &Server{logger: l}
	server.srv, err = tcpserver.NewServer(address)
	server.srv.SetListenConfig(&tcpserver.ListenConfig{
		SocketReusePort: true,
	})
	server.Option(opts...)
	return
}

// Option sets the options specified.
func (s *Server) Option(opts ...ServerOption) {
	for _, opt := range opts {
		opt(s)
	}
}

// WithTimeout sets the timeout for the server. Default is 10 minutes.
func WithTimeout(to time.Duration) ServerOption {
	const defaultTimeout = 10 * time.Minute
	if to == 0 {
		return func(s *Server) {
			s.timeout = defaultTimeout
		}
	}
	return func(s *Server) {
		s.timeout = to
	}
}

func WithSocketReusePort(reuse bool) ServerOption {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketReusePort = reuse
		s.srv.SetListenConfig(c)
	}
}

func WithSocketFastOpen(fopen bool) ServerOption {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketFastOpen = fopen
		s.srv.SetListenConfig(c)
	}
}

func WithSocketDeferAccept(daccept bool) ServerOption {
	return func(s *Server) {
		c := s.srv.GetListenConfig()
		c.SocketDeferAccept = daccept
		s.srv.SetListenConfig(c)
	}
}

func WithLoops(loops int) ServerOption {
	return func(s *Server) {
		s.srv.SetLoops(loops)
	}
}

func WithWorkerpoolShards(shards int) ServerOption {
	return func(s *Server) {
		s.srv.SetWorkerpoolShards(shards)
	}
}

func WithAllowThreadLocking(locking bool) ServerOption {
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
	s.logger = s.logger.With().Str("proto", p.Name()).Logger()
	s.Handler = NewHandler(s.logger, p, s.timeout)
	s.RegisterHandler(s.Handler.Handle)
}

// ListenAndServe starts the TCP server.
func (s *Server) ListenAndServe() error {
	defer func() {
		for _, name := range s.Handler.subsribersNames() {
			e := new(event.GenericEvent)
			e.SetName(fmt.Sprintf("%s.%s", event.CloseConnection, name))
			go s.Handler.fireEvent(context.Background(), &s.logger, e)
		}
		err := s.srv.Shutdown(s.timeout)
		if err != nil {
			return
		}
	}()

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
