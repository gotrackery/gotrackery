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
	cfg := config{
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
		SamplePG: samplePGDatabase{
			URI: "postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable",
		},
	}
	b, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	fmt.Println(string(b))
}

func TestUnmarshalConfig(t *testing.T) {
	txt := []byte(`player:
  address: 100.20.3.44:5678
  proto: egts
  in: /tmp/cap/egts/out
  nums: 1
  timeouts: 1
tcp:
  address: 100.20.3.44:5678
  proto: egts
  socket_reuse_port: true
  socket_fast_open: false
  socket_defer_accept: true
  workerpool_shards: 1
sample_db:
  uri: postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable
`)
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(txt))
	require.NoError(t, err)
	var cfg config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err)
	fmt.Println(cfg)
	fmt.Println(viper.IsSet("sample_db.uri"))
}
