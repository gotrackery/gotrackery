package cfg

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gookit/event"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/protocol/wialonips"
	"github.com/gotrackery/gotrackery/internal/sampledb"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	_ zerolog.LogObjectMarshaler = (*logging)(nil)
	_ zerolog.LogObjectMarshaler = (*player)(nil)
	_ zerolog.LogObjectMarshaler = (*tcpServer)(nil)
	_ zerolog.LogObjectMarshaler = (*consumers)(nil)
)

type logging struct {
	Level   string
	Console bool
	NoBlock bool `mapstructure:"no-block" yaml:"no-block"`
}

func (l logging) MarshalZerologObject(e *zerolog.Event) {
	e.Dict("logging", e.Str("level", l.Level))
}

type player struct {
	Address  string
	Proto    string
	InPath   string `mapstructure:"in" yaml:"in"`
	FileMask string `mapstructure:"mask" yaml:"mask"`
	Workers  int    `mapstructure:"nums" yaml:"nums"`
	Delay    int
	Timeouts int
}

func (p player) MarshalZerologObject(e *zerolog.Event) {
	e.Str("address", p.Address)
	e.Str("proto", p.Proto)
	e.Str("in-path", p.InPath)
	e.Str("file-mask", p.FileMask)
	e.Int("workers", p.Workers)
	e.Int("timeouts", p.Timeouts)
	e.Int("delay", p.Delay)
}

type tcpServer struct {
	Address            string
	Proto              string
	Timeouts           int
	SocketReusePort    bool `mapstructure:"socket-reuse-port" yaml:"socket-reuse-port"`
	SocketFastOpen     bool `mapstructure:"socket-fast-open" yaml:"socket-fast-open"`
	SocketDeferAccept  bool `mapstructure:"socket-defer-accept" yaml:"socket-defer-accept"`
	Loops              int
	WorkerpoolShards   int  `mapstructure:"workerpool-shards" yaml:"workerpool-shards"`
	AllowThreadLocking bool `mapstructure:"allow-thread-locking" yaml:"allow-thread-locking"`
}

func (s tcpServer) MarshalZerologObject(e *zerolog.Event) {
	e.Str("address", s.Address)
	e.Str("proto", s.Proto)
	e.Int("timeouts", s.Timeouts)
	e.Bool("socket-reuse-port", s.SocketReusePort)
	e.Bool("socket-fast-open", s.SocketFastOpen)
	e.Bool("socket-defer-accept", s.SocketDeferAccept)
	e.Int("loops", s.Loops)
	e.Int("workerpool-shards", s.WorkerpoolShards)
	e.Bool("allow-thread-locking", s.AllowThreadLocking)
}

type samplePGDatabase struct {
	URI string
}

type consumers struct {
	SamplePG samplePGDatabase `mapstructure:"sample-db" yaml:"sample-db"`
	// Notifier telegram
}

// type telegram struct {
// 	Token  string
// 	ChatID int
// }

func (c consumers) MarshalZerologObject(e *zerolog.Event) {
	e.Str("sample-db", c.SamplePG.URI)
}

type Config struct {
	Log       logging `mapstructure:"logging" yaml:"logging"`
	Player    player
	TCPServer tcpServer `mapstructure:"tcp" yaml:"tcp"`
	Consumers consumers
}

func Load() (*Config, error) {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &c, nil
}

/* logging methods */

func (l logging) ZerologLevel() zerolog.Level {
	if !viper.IsSet("logging.level") {
		return zerolog.DebugLevel
	}

	switch l.Level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

/* player methods */

func (p player) GetOptions() []tcp.ReplayerOption {
	o := make([]tcp.ReplayerOption, 0, 2)
	if viper.IsSet("player.delay") {
		o = append(o, tcp.WithDelay(p.Delay))
	}
	if viper.IsSet("player.timeouts") {
		o = append(o, tcp.WithTimeouts(time.Duration(p.Timeouts)*time.Second))
	}
	return o
}

// GetProtocol returns tcp.Protocol for bufio.Scanner.
// If protocol is not defined it will return bufio.ScanLines.
func (p player) GetProtocol() tcp.Protocol {
	var splitFuncs = map[string]tcp.Protocol{
		wialonips.Proto: wialonips.NewWialonIPS(),
		egts.Proto:      egts.NewEGTS(),
	}

	if splitFunc, ok := splitFuncs[p.Proto]; ok {
		return splitFunc
	}

	// ToDo return wialon retranslator here as default.
	return nil
}

func (p player) Validate() error {
	err := pathExists(p.InPath)
	if err != nil {
		return fmt.Errorf("validate InPath: %w", err)
	}

	_, err = net.ResolveTCPAddr("tcp", p.Address)
	if err != nil {
		return fmt.Errorf("validate Address: %w", err)
	}
	return nil
}

/* tcp server methods */

// GetProtocol returns protocol handler according Proto property.
func (s tcpServer) GetProtocol() tcp.Protocol {
	switch s.Proto {
	case egts.Proto:
		return egts.NewEGTS()
	case wialonips.Proto:
		return wialonips.NewWialonIPS()
	}
	return egts.NewEGTS()
}

func (s tcpServer) GetOptions() []tcp.ServerOption {
	o := make([]tcp.ServerOption, 0, 2)
	if viper.IsSet("tcp.timeouts") {
		o = append(o, tcp.WithTimeout(time.Duration(s.Timeouts)*time.Second))
	}
	if viper.IsSet("tcp.socket-reuse-port") {
		o = append(o, tcp.WithSocketReusePort(s.SocketReusePort))
	}
	if viper.IsSet("tcp.socket-fast-open") {
		o = append(o, tcp.WithSocketFastOpen(s.SocketFastOpen))
	}
	if viper.IsSet("tcp.socket-defer-accept") {
		o = append(o, tcp.WithSocketDeferAccept(s.SocketDeferAccept))
	}
	if viper.IsSet("tcp.loops") {
		o = append(o, tcp.WithLoops(s.Loops))
	}
	if viper.IsSet("tcp.workerpool-shards") {
		o = append(o, tcp.WithWorkerpoolShards(s.WorkerpoolShards))
	}
	if viper.IsSet("tcp.allow-thread-locking") {
		o = append(o, tcp.WithAllowThreadLocking(s.AllowThreadLocking))
	}
	return o
}

func (s tcpServer) Validate() error {
	_, err := net.ResolveTCPAddr("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("validate Address: %w", err)
	}
	return nil
}

/* consumers methods */

func (c consumers) Subscribers() (subs []event.Subscriber, err error) {
	postgres, err := c.SamplePG.Subscriber()
	if err != nil {
		return nil, err
	}
	/*
		telegram, err := c.Notifier.Subscriber()
		if err != nil {
			return nil, err
		}
	*/
	return append(make([]event.Subscriber, 0), postgres), nil
}

func (s samplePGDatabase) Subscriber() (sub event.Subscriber, err error) {
	if !viper.IsSet("consumers.sample-db.uri") {
		return nil, nil
	}

	var (
		p  *pgxpool.Pool
		db *sampledb.DB
	)
	ctx := context.Background()
	p, err = pgxpool.New(ctx, s.URI)
	if err != nil {
		return nil, fmt.Errorf("connect to sample db: %w", err)
	}
	err = p.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping sample db: %w", err)
	}

	db, err = sampledb.NewDB(p)
	if err != nil {
		return nil, fmt.Errorf("create sample listener: %w", err)
	}
	return db, nil
}

/* utils */

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
