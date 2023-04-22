package cmd

import (
	"fmt"

	"github.com/gotrackery/gotrackery/cfg"
	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tcpCmd represents the tcp command.
var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "tcp command runs server with given protocol.",
	Long: `Run TCP server with given protocol at given port.
 Server will handle incoming telematic data and provide it for subscribers.
 You can replay (see replay command) recorded data from another server.

Example:
 gotr tcp -p egts -a :5001 --sample_db "postgres://postgres:postgres@localhost:5432/sample_db?sslmode=disable"`,
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlag("tcp.address", rootCmd.PersistentFlags().Lookup("address"))
		_ = viper.BindPFlag("tcp.proto", rootCmd.PersistentFlags().Lookup("proto"))
		_ = viper.BindPFlag("tcp.timeouts", rootCmd.PersistentFlags().Lookup("timeouts"))

		viper.SetDefault("tcp.address", ":5001")
		viper.SetDefault("tcp.proto", egts.Proto)
		viper.SetDefault("tcp.timeouts", 10)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cfg.Load()
		if err != nil {
			return fmt.Errorf("loading config in tcp command: %w", err)
		}

		err = c.TCPServer.Validate()
		if err != nil {
			return fmt.Errorf("validate tcp server config: %w", err)
		}

		logger = internal.NewLogger(c.Log.ZerologLevel(), c.Log.Console, c.Log.NoBlock)
		logger.Info().Object("tcp-server", c.TCPServer).Msg("server config")
		logger.Info().Object("consumers", c.Consumers).Msg("consumers config")

		srv, err := tcp.NewServer(logger, c.TCPServer.Address, c.TCPServer.GetOptions()...)
		if err != nil {
			return fmt.Errorf("create tcp server: %w", err)
		}

		srv.SetProtocol(c.TCPServer.GetProtocol())
		cons, err := c.Consumers.GetConsumers(&logger)
		if err != nil {
			return fmt.Errorf("get consumers: %w", err)
		}

		for _, con := range cons {
			srv.Handler.RegisterEventListener(con)
		}
		err = srv.ListenAndServe()
		if err != nil {
			return fmt.Errorf("start tcp server: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tcpCmd)
}
