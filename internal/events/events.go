package events

import (
	"encoding/json"
	"github.com/launchboxio/operator/internal/stream"
)

type ActionCableEvent struct {
	Message ActionCableEventMessage `json:"message"`
}

type ActionCableEventMessage struct {
	Type    string          `json:"type"`
	Id      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type AckEvent struct {
	EventId string `json:"eventId"`
}

func (ack AckEvent) Marshal() ([]byte, error) {
	return json.Marshal(ack)
}

func (acem *ActionCableEventMessage) GetPayload() (map[string]interface{}, error) {
	result := map[string]interface{}{}
	err := json.Unmarshal(acem.Payload, &result)
	return result, err
}

type ActionCableEventIdentifier struct {
	Channel   string `json:"channel"`
	ClusterId int    `json:"cluster_id"`
}

func (ace *ActionCableEvent) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, ace); err != nil {
		return err
	}

	return nil
}

type ProjectStatusEvent struct {
	Status        string `json:"status"`
	CaCertificate string `json:"ca_certificate,omitempty"`
	Action        string `json:"action"`
	ProjectId     int    `json:"project_id"`
}

func NewProjectStatusEvent(projectId int, status string, caCertificate []byte) *ProjectStatusEvent {
	return &ProjectStatusEvent{
		Action:        "project_status",
		ProjectId:     projectId,
		Status:        status,
		CaCertificate: string(caCertificate),
	}
}

func (pse ProjectStatusEvent) Marshal() ([]byte, error) {
	data, err := json.Marshal(pse)
	if err != nil {
		return nil, err
	}
	return stream.BaseEvent{
		Command: "message",
		Data:    string(data),
	}.Marshal()
}

type AckResponse struct {
	EventId string `json:"id"`
}

const (
	ProjectCreatedEvent string = "projects.created"
	ProjectPausedEvent         = "projects.paused"
	ProjectResumedEvent        = "projects.resumed"
	ProjectUpdatedEvent        = "projects.updated"
	ProjectDeletedEvent        = "projects.deleted"
	AddonCreatedEvent          = "addons.created"
	AddonUpdatedEvent          = "addons.update"
	AddonDeletedEvent          = "addons.delete"
	PingEvent                  = "ping"
)

type EventHandler = func(event *ActionCableEvent) error
