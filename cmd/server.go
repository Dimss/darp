package main

import (
	"context"
	"crypto/tls"
	"github.com/darp/pkg/forwarder"
	"github.com/darp/pkg/webhook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var runWebhookServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server for serving DARP requests",
	Run: func(cmd *cobra.Command, args []string) {
		StartHttpRouter()
	},
}

func AddDoneChannelContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add doneChan to context
		ctx := context.WithValue(r.Context(), "doneChan", make(chan forwarder.UpstreamResponse))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func StartHttpRouter() {
	cert := viper.GetString("http.crt")
	key := viper.GetString("http.key")
	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		logrus.Error("Failed to load key pair: %v", err)
	}
	mux := http.NewServeMux()

	// Handel admission validation webhook request
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		webhook.ValidateWebHookHandler(w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		webhook.LivenessHandler(w, r)
	})
	// Create HTTPS server configuration
	s := &http.Server{
		Addr:      "0.0.0.0:8080",
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		Handler:   AddDoneChannelContext(mux),
	}
	logrus.Infof("Starting HTTPS server on %v", s.Addr)
	// Start HTTPS server
	logrus.Fatal(s.ListenAndServeTLS("", ""))

}
