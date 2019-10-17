package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "darp",
	Short: "DARP - Dynamic Admission Reverse Proxy",
}

func init() {
	// Init config
	cobra.OnInitialize(initConfig)
	// Setup commands
	rootCmd.AddCommand(runWebhookServerCmd)
	rootCmd.PersistentFlags().StringP("configpath", "c", "", "Path to config directory with config.json file, default to . ")
	rootCmd.PersistentFlags().BoolP("killswitch", "-i", true, "Default AdmissionReview IsAllowed value on error, default value true")
	if err := viper.BindPFlag("configpath", rootCmd.PersistentFlags().Lookup("configpath")); err != nil {
		panic(err)
	}
	// Init log
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}
func initConfig() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("DARP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	configPath := viper.GetString("configpath")
	logrus.Infof("Configuration directory %v: ", configPath)
	// If config flag is empty, assume config.json located in current directory
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorf("Unable to read config.json file, err: %s", err)
		os.Exit(1)
	}
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
