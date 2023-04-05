package cfg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gookit/event"
	"github.com/gotrackery/gotrackery/internal/protocol/detector"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/protocol/wialonips"
	"github.com/gotrackery/gotrackery/internal/sampledb"
	"github.com/gotrackery/gotrackery/internal/tcp/replayer"
	"github.com/gotrackery/gotrackery/internal/tcp/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type logging struct {
	Level   string
	Console bool
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

type samplePGDatabase struct {
	URI string
}

type consumers struct {
	SamplePG samplePGDatabase `mapstructure:"sample-db" yaml:"sample-db"`
}

type Config struct {
	Log       logging
	Player    player
	TCPServer tcpServer `mapstructure:"tcp" yaml:"tcp"`
	Consumers consumers
}

func Load() (*Config, error) {
	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if err := c.Player.validate(); err != nil {
		return nil, fmt.Errorf("validate player config: %w", err)
	}
	return &c, nil
}

/* logging methods */

func (l logging) ZerologLevel() zerolog.Level {
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

func (p player) GetOptions() []replayer.Option {
	o := make([]replayer.Option, 0, 2)
	if viper.IsSet("player.delay") {
		o = append(o, replayer.WithDelay(p.Delay))
	}
	if viper.IsSet("player.timeouts") {
		o = append(o, replayer.WithTimeouts(time.Duration(p.Timeouts)*time.Second))
	}
	return o
}

// GetSplitFunc returns bufio.SplitFunc for bufio.Scanner.
// If protocol is not defined it will return bufio.ScanLines.
func (p player) GetSplitFunc() bufio.SplitFunc {
	var splitFuncs = map[string]bufio.SplitFunc{
		wialonips.Proto: wialonips.GetSplitFunc(),
		egts.Proto:      egts.GetSplitFunc(),
	}

	if p.Proto == detector.Proto {
		return detector.NewProtocolScanner().ScanProtocol
	}

	if splitFunc, ok := splitFuncs[p.Proto]; ok {
		return splitFunc
	}
	return bufio.ScanLines
}

func (p player) validate() error {
	err := pathExists(p.InPath)
	if err != nil {
		return fmt.Errorf("validate InPath: %w", err)
	}
	return nil
}

/* tcp server methods */

// GetProtocol returns protocol handler according Proto property.
func (s tcpServer) GetProtocol() server.Protocol {
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

func (s tcpServer) GetOptions() []server.Option {
	o := make([]server.Option, 0, 2)
	if viper.IsSet("tcp.timeouts") {
		o = append(o, server.WithTimeout(time.Duration(s.Timeouts)*time.Second))
	}
	if viper.IsSet("tcp.socket-reuse-port") {
		o = append(o, server.WithSocketReusePort(s.SocketReusePort))
	}
	if viper.IsSet("tcp.socket-fast-open") {
		o = append(o, server.WithSocketFastOpen(s.SocketFastOpen))
	}
	if viper.IsSet("tcp.socket-defer-accept") {
		o = append(o, server.WithSocketDeferAccept(s.SocketDeferAccept))
	}
	if viper.IsSet("tcp.loops") {
		o = append(o, server.WithLoops(s.Loops))
	}
	if viper.IsSet("tcp.workerpool-shards") {
		o = append(o, server.WithWorkerpoolShards(s.WorkerpoolShards))
	}
	if viper.IsSet("tcp.allow-thread-locking") {
		o = append(o, server.WithAllowThreadLocking(s.AllowThreadLocking))
	}
	return o
}

/* consumers methods */

func (c consumers) GetConsumers(l *zerolog.Logger) (lsnr []event.Listener, err error) {
	lsnr = make([]event.Listener, 0)
	sampleDB, err := c.SamplePG.GetConsumer(l)
	if err != nil {
		return nil, fmt.Errorf("get sample db consumer: %w", err)
	}
	if sampleDB != nil {
		lsnr = append(lsnr, sampleDB)
	}
	return lsnr, nil
}

func (s samplePGDatabase) GetConsumer(l *zerolog.Logger) (lsnr event.Listener, err error) {
	if !viper.IsSet("sample-db.uri") {
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

	db, err = sampledb.NewDB(l, p)
	if err != nil {
		return nil, fmt.Errorf("create sample listener: %w", err)
	}
	return db, nil
}

/* utils */

func pathExists(path string) error {
	fmt.Println("path:", path)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("invalid path: %s", path)
	}
	return err
}
