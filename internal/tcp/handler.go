package tcp

import (
	"bufio"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gotrackery/gotrackery/internal"
	ev "github.com/gotrackery/gotrackery/internal/event"
	gen "github.com/gotrackery/gotrackery/internal/protocol"
	"github.com/gotrackery/protocol"

	"github.com/gookit/event"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/maurice2k/tcpserver"
	"github.com/rs/zerolog"
)

const (
	evName = "generic.position.receive"
)

// Result is the result of handling bytes packet.
// It contains:
// CloseSession flag that signals that the session should be closed. In case like login message was failed.
// Response is the response bytes that shall be sent to the client.
// GenericAdapter that is used to get extracted and unified data from the packet.
type Result struct {
	CloseSession   bool
	Response       []byte
	GenericAdapter gen.Adapter
}

// Protocol is the tcp protocol contract.
type Protocol interface {
	// GetName returns the name of the protocol.
	GetName() string
	// GetSplitFunc returns the split function for the protocol.
	GetSplitFunc() bufio.SplitFunc
	// Respond returns the Result as passed was processed.
	Respond(*internal.Session, []byte) (Result, error)
}

// Handler is the tcp protocol handler.
type Handler struct {
	logger      zerolog.Logger
	IdleTimeout time.Duration
	proto       Protocol
	evManager   *event.Manager
}

// NewHandler creates a new tcp protocol handler.
func NewHandler(l zerolog.Logger, p Protocol, idle time.Duration) (h *Handler) {
	h = &Handler{logger: l, proto: p, IdleTimeout: idle, evManager: event.NewManager("receiver")}
	return
}

// RegisterEventListener registers an event listener to get a copy of Event that handler extracted.
func (h *Handler) RegisterEventListener(listener event.Listener) {
	h.evManager.AddListener(evName, listener)
}

// Handle handles the tcp inbound connection.
// It parses the tcp payload and calls the protocol handler.
// If some useful Event is extracted from the packet,
// it will be sent to the listeners what previously was registered by calling RegisterEventListener.
func (h *Handler) Handle(conn tcpserver.Connection) {
	session := gonanoid.Must(8)

	logger := h.logger.With().
		Str("session", session).
		Str("remote", conn.RemoteAddr().String()).
		Str("local", conn.LocalAddr().String()).
		Logger()
	logger.Debug().Msg("session opened...")
	defer func() {
		logger.Debug().Dur("opened", time.Since(conn.GetStartTime())).Msg("session closed...")
		err := conn.Close()
		if err != nil {
			logger.Err(err).Msg("close session")
		}
	}()

	h.handle(&logger, conn)
}

func (h *Handler) handle(l *zerolog.Logger, conn tcpserver.Connection) {
	err := conn.SetDeadline(time.Now().Add(h.IdleTimeout))
	if err != nil {
		l.Error().Err(err).Msg("set deadline")
		return
	}

	var result Result
	session := internal.NewSession()

	scanner := bufio.NewScanner(conn)
	scanner.Split(h.proto.GetSplitFunc())
	for scanner.Scan() {
		if errors.Is(scanner.Err(), protocol.ErrInconsistentData) {
			l.Error().Err(scanner.Err()).Str("bytes", hex.EncodeToString(scanner.Bytes())).Msg("scan payload")
			return
		}

		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		err = conn.SetDeadline(time.Now().Add(h.IdleTimeout))
		if err != nil {
			l.Error().Err(err).Msg("extending deadline")
			return
		}

		l.Debug().Str("dir", "in").Str("bytes", hex.EncodeToString(data)).Send()
		result, err = h.proto.Respond(session, data)
		if err != nil {
			l.Warn().Err(err).Msg("got protocol error")
		}
		l.Debug().Str("dir", "out").Str("device", session.GetDevice()).Str("bytes", hex.EncodeToString(result.Response)).Send()

		if len(result.Response) > 0 {
			_, err = conn.Write(result.Response) // send result even got error
			if err != nil {
				l.Error().Err(err).Msg("write result")
				return
			}
		}

		if result.CloseSession {
			return
		}

		if result.GenericAdapter != nil {
			err = h.fireEvent(result.GenericAdapter)
			if err != nil {
				l.Error().Err(err).Msg("fire event")
			}
		}
	}
}

func (h *Handler) fireEvent(adapter gen.Adapter) error {
	poss := adapter.GenericPositions()
	for _, pos := range poss {
		e := &ev.GenericEvent{
			Position: pos,
		}
		e.SetName(evName)
		err := h.evManager.FireEvent(e)
		if err != nil {
			// ToDo log every error separately
			return err
		}
		h.logger.Debug().Str("event", evName).Object("position", pos).Msg("event fired")
	}

	return nil
}
