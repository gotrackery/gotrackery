package cfg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gookit/event"
	"github.com/gotrackery/gotrackery/internal/protocol/detector"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/protocol/wialonips"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/gotrackery/gotrackery/internal/traccar"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

var splitFuncs = map[string]bufio.SplitFunc{
	wialonips.Proto: wialonips.GetSplitFunc(),
	egts.Proto:      egts.GetSplitFunc(),
}

// Player - config for playing previously recorded data.
type Player struct {
	Address   string
	Proto     string
	InPath    string
	FileMask  string
	Workers   int
	Delay     int
	Timeouts  time.Duration
	splitFunc bufio.SplitFunc
}

// Load config values from pflag.FlagSet.
func (p *Player) Load(f *pflag.FlagSet) (err error) {
	p.Address, err = f.GetString("address")
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}
	p.Proto, err = f.GetString("proto")
	if err != nil {
		return fmt.Errorf("failed to get proto: %w", err)
	}
	p.InPath, err = f.GetString("in")
	if err != nil {
		return fmt.Errorf("failed to get in: %w", err)
	}
	p.FileMask, err = f.GetString("mask")
	if err != nil {
		return fmt.Errorf("failed to get mask: %w", err)
	}
	p.Workers, err = f.GetInt("nums")
	if err != nil {
		return fmt.Errorf("failed to get nums: %w", err)
	}
	if t, err := f.GetInt("timeouts"); err != nil {
		return fmt.Errorf("failed to get timeouts: %w", err)
	} else {
		p.Timeouts = time.Duration(t) * time.Second
	}
	p.Delay, err = f.GetInt("delay")
	if err != nil {
		return fmt.Errorf("failed to get delay: %w", err)
	}

	if p.Proto == detector.Proto {
		p.splitFunc = detector.NewProtocolScanner().ScanProtocol
		return nil
	}

	var ok bool
	if p.splitFunc, ok = splitFuncs[p.Proto]; !ok {
		return errors.New("invalid protocol")
	}
	return nil
}

func (p *Player) Validate() error {
	err := pathExists(p.InPath)
	if err != nil {
		return fmt.Errorf("invalid in path: %w", err)
	}
	return nil
}

func pathExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("invalid path: %s", path)
	}
	return err

}

// GetSplitFunc returns function for splitting data for emulating sending data by packages.
func (p *Player) GetSplitFunc() bufio.SplitFunc {
	return p.splitFunc
}

// TCPServer - config for TCP server that receives and handles TCP protocols.
type TCPServer struct {
	Address   string
	Proto     string
	TraccarDB string
}

// Load config values from pflag.FlagSet.
func (s *TCPServer) Load(f *pflag.FlagSet) (err error) {
	s.Address, err = f.GetString("address")
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}
	s.Proto, err = f.GetString("proto")
	if err != nil {
		return fmt.Errorf("failed to get proto: %w", err)
	}
	s.TraccarDB, err = f.GetString("traccar")
	if err != nil {
		return fmt.Errorf("failed to get traccar db: %w", err)
	}
	return nil
}

// GetProtocol returns protocol handler according Proto property.
func (s *TCPServer) GetProtocol() tcp.Protocol {
	switch s.Proto {
	case detector.Proto:
		return detector.NewDetector([]byte{})
	case egts.Proto:
		return egts.NewEGTS()
	case wialonips.Proto:
		return wialonips.NewWialonIPS()
	}
	return egts.NewEGTS()
}

// GetConsumers returns slice of event listeners to consume parsed and unified data that received by server.
func (s *TCPServer) GetConsumers(l *zerolog.Logger) (listeners []event.Listener, err error) {
	if s.TraccarDB == "" {
		return
	}

	var (
		p  *pgxpool.Pool
		db *traccar.DB
	)
	p, err = pgxpool.New(context.Background(), s.TraccarDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to traccar db: %w", err)
	}

	db, err = traccar.NewDB(l, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create traccar listener: %w", err)
	}
	listeners = append(listeners, db)
	return listeners, nil
}
