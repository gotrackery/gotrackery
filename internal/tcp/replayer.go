package tcp

import (
	"bufio"
	"encoding/hex"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/gotrackery/gotrackery/internal/replayer"
	"github.com/rs/zerolog/log"
)

var _ replayer.Replayer = (*Replayer)(nil)

// Replayer is replayer previously recorded data of TCP protocols.
type Replayer struct {
	splitter     bufio.SplitFunc
	addr         string
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	packetDelay  int
}

// NewReplayer creates new instance of Replayer.
// Provide server address to where data will be sent and splitter to send package by package with delay between.
// It is possible but not recommended to use splitter by EOF to send whole file as one package.
func NewReplayer(address string, splitter bufio.SplitFunc) *Replayer {
	return &Replayer{
		addr:         address,
		splitter:     splitter,
		dialTimeout:  10 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		packetDelay:  100,
	}
}

func (p *Replayer) SetTimeouts(to time.Duration) *Replayer {
	p.dialTimeout = to
	p.readTimeout = to
	p.writeTimeout = to
	return p
}

func (p *Replayer) SetDelay(milsecs int) *Replayer {
	p.packetDelay = milsecs
	return p
}

// Replay sends data from a file with given filename to a server.
func (p *Replayer) Replay(filename string) error {
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
	scanner.Split(p.splitter)
	for scanner.Scan() {
		t := time.Now()
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
	scanner.Split(p.splitter)
	for scanner.Scan() {
		resp = scanner.Bytes()
		if len(resp) == 0 {
			continue
		}
		return resp, nil
	}
	return nil, scanner.Err()
}
