package cmd

import (
	"bytes"
	"fmt"
	"github.com/colindr/gotests/gosync/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var host string
var port int
var configFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "host to listen on")
	rootCmd.PersistentFlags().IntVar(&port, "port", 0, "port the daemon should listen on")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")

	// TODO: add http port for http REST API
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.SetDefault("port", 4200)

	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.SetDefault("host", "0.0.0.0")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {

		// Search config in home directory with name "gosync.{yml,yaml}"
		viper.AddConfigPath("/etc/gosyncd")
		viper.SetConfigName("gosyncd")
	}

	if err := viper.ReadInConfig(); err != nil {
		empty := []byte(``)

		if err := viper.ReadConfig(bytes.NewBuffer(empty)); err != nil {
			fmt.Println("Error with fake empty config...")
		}

		fmt.Println("No gosync config, using defaults")
	}
}

var rootCmd = &cobra.Command{
	Use:   "gosyncd",
	Short: "gosyncd is the gosync daemon",
	Long:  `An exercise in golang to sync files`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the gosync daemon
		daemon.StartDaemon()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
