package cmd

import (
	"fmt"

	"github.com/gotrackery/gotrackery/cfg"
	"github.com/gotrackery/gotrackery/internal/player"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// replayCmd represents the replay command.
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Play telematics data collected with tcpdump and extracted payload with tcpflow",
	Long: ` Run tcpdump against working endpoint and save a dump to file.
 Extract payload from saved dump file with tcpflow into separate dir.
 Use replay command to replay extracted payload to given sever.

 Example:
 sudo tcpdump -w dump.pcap -i <network interface> port <port>
 sudo tcpflow -r dump.pcap -o ./dump
 gotr replay -a :5001 -p egts -i ./dump"
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlag("player.address", rootCmd.PersistentFlags().Lookup("address"))
		_ = viper.BindPFlag("player.proto", rootCmd.PersistentFlags().Lookup("proto"))
		_ = viper.BindPFlag("player.timeouts", rootCmd.PersistentFlags().Lookup("timeouts"))
		_ = viper.BindPFlag("player.in", rootCmd.PersistentFlags().Lookup("in"))
		_ = viper.BindPFlag("player.mask", rootCmd.PersistentFlags().Lookup("mask"))
		_ = viper.BindPFlag("player.nums", rootCmd.PersistentFlags().Lookup("nums"))
		_ = viper.BindPFlag("player.delay", rootCmd.PersistentFlags().Lookup("delay"))

		viper.SetDefault("player.address", ":5001")
		viper.SetDefault("player.proto", egts.Proto)
		viper.SetDefault("player.timeouts", 10)
		viper.SetDefault("player.in", "./in")
		viper.SetDefault("player.mask", "*")
		viper.SetDefault("player.nums", 200)
		viper.SetDefault("player.delay", 100)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cfg.Load()
		if err != nil {
			return fmt.Errorf("loading config in player command: %w", err)
		}
		if err := c.Player.Validate(); err != nil {
			return fmt.Errorf("validate player config: %w", err)
		}

		logger.Info().Object("player", c.Player).Msg("player config")

		replayer := tcp.NewReplayer(
			c.Player.Address,
			c.Player.Protocol(),
			c.Player.Options()...,
		)
		player.Run(c.Player.InPath, c.Player.FileMask, replayer, c.Player.Workers)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)

	rootCmd.PersistentFlags().StringP("in", "i", "./in",
		"path to input files: -i ./in_files")
	rootCmd.PersistentFlags().StringP("mask", "m", "*", "file mask: -m *.csv")
	rootCmd.PersistentFlags().IntP("nums", "n", 200,
		"number of emulating devices: -n 10")
	rootCmd.PersistentFlags().IntP("delay", "d", 100,
		"max random delay between sending packets in milliseconds: -d 100")
}
