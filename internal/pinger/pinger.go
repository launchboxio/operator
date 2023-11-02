package pinger

import (
	"github.com/go-logr/logr"
	"github.com/launchboxio/operator/internal/client"
	"time"
)

type Pinger struct {
	Client *client.Client
	Logger logr.Logger
}

func New(httpClient *client.Client, logger logr.Logger) *Pinger {
	return &Pinger{
		Client: httpClient,
		Logger: logger,
	}
}

func (p *Pinger) Start(clusterId int) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	operator := p.Client.Operator()

	data := &client.PingRequest{
		Cluster: client.ClusterPing{
			Version:         "1.25.15",
			AgentVersion:    "1.2.3",
			Provider:        "launchbox",
			Region:          "us-east-1",
			AgentIdentifier: "localhost",
		},
	}
	if _, err := operator.Ping(clusterId, data); err != nil {
		p.Logger.Error(err, "Failed to ping HQ")
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				if _, err := operator.Ping(clusterId, data); err != nil {
					p.Logger.Error(err, "Failed to ping HQ")
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
