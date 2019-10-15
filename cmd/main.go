package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "secroute",
	Short: "SecRoute Webhook - enforce creation of secure routes",
}



func init() {
	// Init config
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "Path to kubeconfig file, default to $home/.kube/config")
	if err := viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig")); err != nil {
		panic(err)
	}
	// Setup commands
	rootCmd.AddCommand(runWebhookServerCmd)
	// Init log
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}
func initConfig() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("DAP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	kubeconfig := viper.GetString("kubeconfig")
	if kubeconfig == "" {
		// Check if kubeconfig file exists in user's HOME
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		_, err := os.Stat(kubeconfig)
		if os.IsNotExist(err) {
			// The kubeconfig wasn't passed in and not found under user's home directory, assuming inClusterConfig mode
			logrus.Info("Unable to find kubeconfig, assuming running inside K8S cluster, gonna use inClusterConfig")
			viper.Set("kubeconfig", "useInClusterConfig")
		} else {
			// Use kubeconfig from user's home directory
			logrus.Info("Gonna use kubeconfig from user's HOME directory")
			viper.Set("kubeconfig", kubeconfig)
		}
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
