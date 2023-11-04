package client

import (
	"fmt"
	"github.com/dghubble/sling"
)

type Operator struct {
	sling *sling.Sling
}

func (c *Client) Operator() *Operator {
	return &Operator{sling: c.sling}
}

type PingRequest struct {
	Cluster ClusterPing `json:"cluster"`
}

type ClusterPing struct {
	Version         string `json:"version"`
	Region          string `json:"region"`
	Provider        string `json:"provider"`
	AgentIdentifier string `json:"agent_identifier"`
	AgentVersion    string `json:"agent_version"`
}
type PingResponse struct{}

func (o *Operator) Ping(clusterId int, data *PingRequest) (*PingResponse, error) {
	path := fmt.Sprintf("/api/v1/clusters/%d/ping", clusterId)
	res := new(PingResponse)
	_, err := o.sling.Post(path).BodyJSON(data).ReceiveSuccess(res)
	return res, err
}
