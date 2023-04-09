package cfg

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshalConfig(t *testing.T) {
	cfg := Config{
		Player: player{
			Address:  "100.20.3.44:5678",
			Proto:    egts.Proto,
			InPath:   "/tmp/cap/egts/out",
			FileMask: "*.data",
			Workers:  1,
			Delay:    1,
			Timeouts: 1,
		},
		TCPServer: tcpServer{
			Address:            "100.20.3.44:5678",
			Proto:              egts.Proto,
			SocketReusePort:    true,
			SocketFastOpen:     false,
			SocketDeferAccept:  true,
			Loops:              1,
			WorkerpoolShards:   1,
			AllowThreadLocking: false,
		},
		Log: logging{
			Level:   "info",
			Console: false,
		},
		Consumers: consumers{
			SamplePG: samplePGDatabase{
				URI: "postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable",
			},
		},
	}
	b, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	fmt.Sprintln(string(b))
}

func TestUnmarshalConfig(t *testing.T) {
	txt := []byte(`log:
    level: info
    console: false
player:
    address: 100.20.3.44:5678
    proto: egts
    in: /tmp/cap/egts/out
    mask: '*.data'
    nums: 1
    delay: 1
    timeouts: 1
tcp:
    address: 100.20.3.44:5678
    proto: egts
    timeouts: 0
    socket-reuse-port: true
    socket-fast-open: false
    socket-defer-accept: false
    loops: 1
    workerpool-shards: 1
    allow-thread-locking: false
consumers:
    sample-db:
        uri: postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable
`)
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(txt))
	require.NoError(t, err)
	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err)
}
