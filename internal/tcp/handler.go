package tcp

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gookit/event"
	"github.com/gotrackery/gotrackery/internal"
	ev "github.com/gotrackery/gotrackery/internal/event"
	gen "github.com/gotrackery/gotrackery/internal/protocol"
	"github.com/gotrackery/protocol/common"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/maurice2k/tcpserver"
	"github.com/rs/zerolog"
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

// Handler is the tcp protocol handler.
type Handler struct {
	logger      zerolog.Logger
	IdleTimeout time.Duration
	proto       Protocol
	evManager   *event.Manager
}

const (
	eventManagerName = "events manager"
)

// NewHandler creates a new tcp protocol handler.
func NewHandler(l zerolog.Logger, p Protocol, idle time.Duration) (h *Handler) {
	h = &Handler{logger: l, proto: p, IdleTimeout: idle, evManager: event.NewManager(eventManagerName)}
	return
}

func (h *Handler) RegisterEventSubscriber(sub event.Subscriber) {
	h.logger.Debug().Str("consumer-listener", fmt.Sprintf("%s", sub)).Msg("register events subscriber")
	h.evManager.AddSubscriber(sub)
}

func (h *Handler) subsribersNames() []string {
	events := h.evManager.ListenedNames()
	set := make(map[string]struct{}, len(events))
	for n := range events {
		names := strings.Split(n, ".")
		name := names[len(names)-1]
		if _, ok := set[name]; !ok {
			set[name] = struct{}{}
		}
	}
	listeners := make([]string, 0)
	for n := range set {
		listeners = append(listeners, n)
	}
	return listeners
}

// Handle handles the tcp inbound connection.
// It parses the tcp payload and calls the protocol handler.
// If some useful Event is extracted from the packet,
// it will be sent to the listeners what previously was registered by calling RegisterEventSubscriber.
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
	splitter := h.proto.NewFrameSplitter()
	scanner.Split(splitter.Splitter())
	for scanner.Scan() {
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
		l.Debug().Str("dir", "out").Str("device", session.Device()).Str("bytes", hex.EncodeToString(result.Response)).Send()

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
			poss := result.GenericAdapter.GenericPositions()
			for _, pos := range poss {
				for _, name := range h.subsribersNames() {
					e := new(ev.GenericEvent)
					e.SetPosition(pos)
					e.SetName(fmt.Sprintf("%s.%s", ev.PositionReceived, name))

					ctx, cancel := context.WithTimeout(context.Background(), h.IdleTimeout)
					defer cancel()
					go h.fireEvent(ctx, l, e)
				}

			}
		}
	}

	if splitter.Error() != nil && errors.Is(splitter.Error(), common.ErrBadData) {
		l.Error().Err(splitter.Error()).Str("bytes", hex.EncodeToString(splitter.BadData())).Msg("bad data")
		return
	}

	if scanner.Err() != nil && scanner.Err() != io.EOF {
		l.Error().Err(scanner.Err()).Msg("scanner error")
		return
	}
}

func (h *Handler) fireEvent(ctx context.Context, l *zerolog.Logger, event event.Event) {
	reply := make(chan ev.Reply)
	defer close(reply)
	retrying := h.retryFire(ctx, h.evManager.FireEvent, reply)

	go func() {
		err := retrying(event)
		if err != nil {
			l.Err(err).Msg("retrying to fire event")
		}
	}()

	for r := range reply {
		if r.Error == nil {
			l.Info().Msg(r.Message)
			return
		}
		l.Err(r.Error).Msg(r.Message)
		if r.Error != nil && r.Error == context.Canceled {
			return
		}
		continue
	}
}

type fireraiser func(event event.Event) error

func (h *Handler) retryFire(ctx context.Context, fireraiser fireraiser, reply chan ev.Reply) fireraiser {
	return func(evnt event.Event) error {
		for try := 1; ; try++ {
			err := fireraiser(evnt)
			if err == nil {
				reply <- ev.Reply{
					Message: fmt.Sprintf(`successful handle "%s" event`, evnt.Name()),
				}
				return nil
			}

			/*
				info: in this place, by creating a new event through the fireraiser, you can call the event of some notifier to notify about an error

				go func(){
					for _, name := range h.subsribersNames() {
						e := new(ev.GenericEvent)
						e.SetName(fmt.Sprintf("%s.%s", ev.NotifyError, name))
						e.SetData(event.M{"message": fmt.Sprintf(`event "%s" failed attempt #[%d]`, evnt.Name(), try)})
						if err := fireraiser(e); err != nil {
							l.Err(err).Msg("notify error")
						}
					}

				}() */

			delay := time.Second << uint(try)
			timer := time.NewTimer(delay)
			defer timer.Stop()
			reply <- ev.Reply{
				Error:   err,
				Message: fmt.Sprintf(`event "%s" failed attempt #[%d] | retrying after: [%s]`, evnt.Name(), try, delay),
			}
			select {
			case <-timer.C:
			case <-ctx.Done():
				reply <- ev.Reply{
					Error:   ctx.Err(),
					Message: fmt.Sprintf(`retrying event "%s" stopped via context`, evnt.Name()),
				}
				return ctx.Err()
			}
		}
	}
}
