package client

import (
	"context"
	"github.com/dghubble/sling"
	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	sling *sling.Sling
}

func New(url string, config clientcredentials.Config) *Client {
	httpClient := config.Client(context.TODO())
	return &Client{
		sling: sling.New().Client(httpClient).Base(url),
	}
}
