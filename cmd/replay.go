package cmd

import (
	"fmt"

	"github.com/gotrackery/gotrackery/cfg"
	"github.com/gotrackery/gotrackery/internal/replayer"
	"github.com/gotrackery/gotrackery/internal/tcp"
	"github.com/spf13/cobra"
)

// replayCmd represents the replay command.
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay telematics data collected with tcpdump and extracted payload with tcpflow",
	Long: ` Run tcpdump against working endpoint and save a dump to file.
 Extract payload from saved dump file with tcpflow into separate dir.
 Use replay command to replay extracted payload to given sever.

 Example:
 sudo tcpdump -w dump.pcap -i <network interface> port <port>
 sudo tcpflow -r dump.pcap -o ./dump
 gotr replay -a :5001 -p egts -i ./dump"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var c cfg.Player
		err := c.Load(cmd.Flags())
		if err != nil {
			return fmt.Errorf("failed to load player config: %w", err)
		}
		err = c.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate player config: %w", err)
		}
		player := tcp.NewReplayer(c.Address, c.GetSplitFunc()).SetTimeouts(c.Timeouts).SetDelay(c.Delay)
		replayer.Run(c.InPath, c.FileMask, player, c.Workers)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)

	replayCmd.PersistentFlags().StringP("in", "i", "./in",
		"path to input files: -i ./in_files")
	replayCmd.PersistentFlags().StringP("mask", "m", "*", "file mask: -m *.csv")
	replayCmd.PersistentFlags().IntP("nums", "n", 200,
		"number of emulating devices: -n 10")
	replayCmd.PersistentFlags().IntP("timeouts", "t", 10,
		"network timeouts in seconds: -t 10")
	replayCmd.PersistentFlags().IntP("delay", "d", 100,
		"max random delay between sending packets in milliseconds: -d 100")
}
