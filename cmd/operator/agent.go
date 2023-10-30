package main

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/launchboxio/operator/internal/events"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/clientcredentials"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"net/url"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Integration Agent",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.New()
		// Start a listener to handle events from the LaunchboxHQ SSE Stream
		streamUrl, _ := cmd.Flags().GetString("stream-url")
		clientId, _ := cmd.Flags().GetString("client-id")
		clientSecret, _ := cmd.Flags().GetString("client-secret")
		tokenUrl, _ := cmd.Flags().GetString("token-url")

		oauthConfig := clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     tokenUrl,
		}

		token, err := oauthConfig.Token(context.TODO())
		if err != nil {
			logger.Error(err, "Failed to get API Token")
			os.Exit(1)
		}

		u, err := url.Parse(streamUrl)
		if err != nil {
			logger.Error(err, "Failed parsing stream URL")
			os.Exit(1)
		}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
			"Authorization": []string{"Bearer " + token.AccessToken},
		})
		if err != nil {
			logger.Error(err, "Failed connecting to Stream URL")
			os.Exit(1)
		}
		defer c.Close()

		kubeClient, err := client.New(config.GetConfigOrDie(), client.Options{})
		utilruntime.Must(corev1alpha1.AddToScheme(kubeClient.Scheme()))

		eventHandler := events.NewHandler(c, zap.New(), kubeClient)
		//go func() {
		logger.Info("Starting event handler")
		_ = eventHandler.Listen(context.TODO())
		//}()
		//go func() {
		//	logger.Info("Starting watcher process")
		//	_ = eventHandler.Watch(context.TODO())
		//}()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().String("stream-url", "ws://localhost:3000/cable", "URL to subscribe for events")
	agentCmd.Flags().String("client-id", "_8ba-5BLULMTuDyl7_Q4bXjkz0BhOJVbzBIVYrDeE5U", "OIDC Client ID")
	agentCmd.Flags().String("client-secret", "yTuR_3p7cxntZnoXOyZqn9lUHdc2_Pooy3gSNcVW1fw", "OIDC Client Secret")
	agentCmd.Flags().String("token-url", "http://localhost:3000/oauth/token", "OAuth token URL")
}
