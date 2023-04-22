package tcp

import (
	"bufio"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/gotrackery/gotrackery/internal/player"
	"github.com/gotrackery/protocol/common"
	"github.com/rs/zerolog/log"
)

const defaultAddr = ":5000"

var _ player.Player = (*Replayer)(nil)

type ReplayerOption func(*Replayer)

// Replayer is replayer previously recorded data of TCP protocols.
type Replayer struct {
	proto        Protocol
	addr         string
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	packetDelay  int
}

// NewReplayer creates new instance of Replayer.
// Provide server address to where data will be sent and proto to send package by package with delay between.
// It is possible but not recommended to use proto by EOF to send whole file as one package.
func NewReplayer(address string, proto Protocol, opts ...ReplayerOption) *Replayer {
	if address == "" {
		address = defaultAddr
	}
	r := &Replayer{
		addr:         address,
		proto:        proto,
		dialTimeout:  10 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		packetDelay:  100,
	}
	r.Option(opts...)
	return r
}

// Option sets the options specified.
func (p *Replayer) Option(opts ...ReplayerOption) {
	for _, opt := range opts {
		opt(p)
	}
}

func WithTimeouts(to time.Duration) ReplayerOption {
	return func(r *Replayer) {
		r.dialTimeout = to
		r.readTimeout = to
		r.writeTimeout = to
	}
}

func WithDelay(milsecs int) ReplayerOption {
	return func(r *Replayer) {
		r.packetDelay = milsecs
	}
}

// Play sends data from a file with given filename to a server.
// ToDo add ctx to break loop
func (p *Replayer) Play(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("can't open file")
		return nil
	}
	defer func() {
		if e := file.Close(); e != nil {
			log.Error().Err(e).Str("filename", filename).Msg("can't close file")
		}
	}()

	d := net.Dialer{Timeout: p.dialTimeout}

	conn, err := d.Dial("tcp", p.addr)
	if err != nil {
		log.Error().Err(err).Str("filename", filename).Msg("can't dial tcp")
		return nil
	}
	defer func() {
		e := conn.Close()
		if e != nil {
			log.Error().Err(e).Str("filename", filename).Msg("can't close tcp connection")
		}
	}()

	log.Info().Str("filename", filename).Msg("replaying")

	scanner := bufio.NewScanner(file)
	splitter := p.proto.NewFrameSplitter()
	scanner.Split(splitter.Splitter())
	for scanner.Scan() {
		t := time.Now()
		if errors.Is(splitter.Error(), common.ErrBadData) {
			log.Error().Err(scanner.Err()).Str("bytes", hex.EncodeToString(splitter.BadData())).Msg("bad data")
			// ToDo add an option to send any data.
			return nil
		}
		b := scanner.Bytes()
		log.Debug().Str("payload", hex.EncodeToString(b)).Msg("sending")
		if err = conn.SetWriteDeadline(time.Now().Add(p.writeTimeout)); err != nil {
			return err
		}
		_, err = conn.Write(scanner.Bytes())
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("writing")
			return nil
		}

		if err = conn.SetReadDeadline(time.Now().Add(p.readTimeout)); err != nil {
			return err
		}
		resp, err := p.readResponse(conn)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("reading")
			return nil
		}
		log.Info().
			Str("filename", filename).
			Str("response", hex.EncodeToString(resp)).
			Dur("elapsed", time.Since(t)).
			Msg("got reply")
		time.Sleep(time.Duration(rand.Intn(p.packetDelay)) * time.Millisecond)
	}
	return nil
}

func (p *Replayer) readResponse(r io.Reader) ([]byte, error) {
	var resp []byte
	scanner := bufio.NewScanner(r)
	splitter := p.proto.NewFrameSplitter()
	scanner.Split(splitter.Splitter())
	for scanner.Scan() {
		if errors.Is(splitter.Error(), common.ErrBadData) {
			return splitter.BadData(), nil
		}

		resp = scanner.Bytes()
		if len(resp) == 0 {
			continue
		}
		return resp, nil
	}
	return nil, scanner.Err()
}
