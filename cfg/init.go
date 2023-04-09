package cfg

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitFile(binary, cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath("./")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", binary))
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msgf("Using config file: %s", viper.ConfigFileUsed())
		if err != nil {
			log.Error().Err(err).Msg("can't write to stderr")
			return
		}
	}
}

func InitConsul(consulAddr, consulKey string) {
	err := viper.AddRemoteProvider("consul", consulAddr, consulKey)
	cobra.CheckErr(err)
	viper.SetConfigType("json")
	err = viper.ReadRemoteConfig()
	cobra.CheckErr(err)
}
