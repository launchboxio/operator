package stream

import "encoding/json"

type Event interface {
	Marshal() ([]byte, error)
}

type BaseEvent struct {
	Data interface{} `json:"data"`
}

func (be BaseEvent) Marshal() ([]byte, error) {
	return json.Marshal(be)
}

type SubscriptionEvent struct {
	Channel   string
	ClusterId int
}

func (se SubscriptionEvent) Marshal() ([]byte, error) {
	id := &struct {
		Channel   string `json:"channel"`
		ClusterId int    `json:"cluster_id"`
	}{
		Channel:   se.Channel,
		ClusterId: se.ClusterId,
	}
	identBytes, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	sub := &struct {
		Command    string `json:"command"`
		Identifier string `json:"identifier"`
	}{
		Command:    "subscribe",
		Identifier: string(identBytes),
	}
	return json.Marshal(sub)
}
