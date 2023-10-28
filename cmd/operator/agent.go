package main

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/launchboxio/operator/internal/events"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Integration Agent",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.New()
		// Start a listener to handle events from the LaunchboxHQ SSE Stream
		streamUrl, _ := cmd.Flags().GetString("stream-url")
		u, err := url.Parse(streamUrl)
		if err != nil {
			logger.Error(err, "Failed parsing stream URL")
			os.Exit(1)
		}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			logger.Error(err, "Failed connecting to Stream URL")
			os.Exit(1)
		}
		defer c.Close()

		kubeClient, err := client.New(config.GetConfigOrDie(), client.Options{})

		eventHandler := events.NewHandler(c, zap.New(), kubeClient)
		go func() {
			logger.Info("Starting event handler")
			_ = eventHandler.Listen(context.TODO())
		}()
		//go func() {
		//	logger.Info("Starting watcher process")
		//	_ = eventHandler.Watch(context.TODO())
		//}()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().String("stream-url", "https://launchboxhq.io/events", "URL to subscribe for events")
}
