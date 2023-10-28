package events

type Event struct {
	Id      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
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
)

type EventHandler = func(event Event) error
