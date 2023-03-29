package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/gotrackery/gotrackery/internal"
	"github.com/gotrackery/gotrackery/internal/protocol/egts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "v0.0.1"
	binary  = "gotr"
	cfgFile string
	logger  zerolog.Logger
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     binary,
	Version: version,
	Short:   "Telematic aggregator service",
	Long: ` Telematics data aggregation server. Designed to receive data from various telematics
data transfer protocols and convert it into a universal stream of messages, to which
you can create a subscriber for further processing of this data, such as saving it to
the database or transferring it to the message queue.

 One running instance of the server processes one type of protocol on a specified port.
 To process multiple protocols, you need to run multiple server instances with different
protocols, which are specified as a command.

Supported commands (protocols):
- tcp
- replay

Supported Subscribers:
- traccar postgres database
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s (c) Copyright 2023 %s\n", binary, viper.GetString("author")) //nolint:forbidigo
		fmt.Printf("Version: %s\n\n", version)                                      //nolint:forbidigo
		return cmd.Help()                                                           //nolint:wrapcheck
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/.%s.yaml)", binary))
	rootCmd.PersistentFlags().StringP("address", "a", ":5001",
		"server address with port: :5001")
	rootCmd.PersistentFlags().StringP("proto", "p", egts.Proto,
		"device protocol to use: egts")
	rootCmd.PersistentFlags().String("traccar",
		"",
		"postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable")

	_ = viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.SetDefault("author", "GoTrackery Authors")

	logger = internal.NewLogger(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Caller().Stack().Logger()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".<binary>" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", binary))
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, err = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		if err != nil {
			log.Error().Err(err).Msg("can't write to stderr")
			return
		}
	}
}
