package main

import (
	"crypto/tls"
	"github.com/darp/pkg/webhook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var runWebhookServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server for serving DAP requests",
	Run: func(cmd *cobra.Command, args []string) {
		StartHttpRouter()
	},
}

func StartHttpRouter() {
	cert := viper.GetString("http.crt")
	key := viper.GetString("http.key")
	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		logrus.Error("Failed to load key pair: %v", err)
	}

	// Handel admission validation webhook request
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		webhook.ValidateWebHookHandler(w, r)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		webhook.LivenessHandler(w, r)
	})
	// Create HTTPS server configuration
	s := &http.Server{
		Addr:      "0.0.0.0:8080",
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}

	logrus.Infof("Starting HTTPS server on %v", s.Addr)
	// Start HTTPS server
	logrus.Fatal(s.ListenAndServeTLS("", ""))

}
