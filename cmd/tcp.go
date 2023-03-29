package cmd

import (
	"fmt"

	"github.com/gotrackery/gotrackery/cfg"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/spf13/cobra"
)

// tcpCmd represents the tcp command.
var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "tcp command runs server with given protocol.",
	Long: `Run TCP server with given protocol at given port.
 Server will handle incoming telematic data and provide it for subscribers.
 You can replay (see replay command) recorded data from another server.

Example:
 gotr tcp -p egts -a :5001 --traccar "postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var c cfg.TCPServer
		err := c.Load(cmd.Flags())
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		srv, err := tcp.NewServer(logger, c.Address)
		if err != nil {
			return fmt.Errorf("failed to create tcp server: %w", err)
		}

		srv.SetProtocol(c.GetProtocol())
		cons, err := c.GetConsumers(&logger)
		if err != nil {
			return fmt.Errorf("failed to get consumers: %w", err)
		}

		for _, con := range cons {
			srv.Handler.RegisterEventListener(con)
		}
		err = srv.ListenAndServe()
		if err != nil {
			return fmt.Errorf("failed to start tcp server: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tcpCmd)
}
