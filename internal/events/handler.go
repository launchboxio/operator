package events

import (
	"encoding/json"
	"github.com/go-logr/logr"
	action_cable "github.com/launchboxio/action-cable"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	Logger logr.Logger
	Client client.Client
}

func New(logger logr.Logger, client client.Client) *Handler {
	handler := &Handler{
		Logger: logger,
		Client: client,
	}
	return handler
}

type HandlerFunc func(event *LaunchboxEvent) error

func (h *Handler) RegisterSubscriptions(stream *action_cable.Stream, identifier map[string]string) {
	subscription := action_cable.NewSubscription(identifier)

	projectHandler := h.projectHandler()
	addonHandler := h.addonHandler()

	subscription.Handler(func(event *action_cable.ActionCableEvent) {
		var handler HandlerFunc
		parsedEvent := &LaunchboxEvent{}
		err := json.Unmarshal(event.Message, parsedEvent)
		if err != nil {
			h.Logger.Error(err, "Failed parsing event")
		}
		switch parsedEvent.Type {
		case "projects.created":
			handler = projectHandler.Create
		case "projects.updated":
			handler = projectHandler.Update
		case "projects.deleted":
			handler = projectHandler.Delete
		case "projects.paused":
			handler = projectHandler.Pause
		case "projects.resumed":
			handler = projectHandler.Resume
		case "addons.created":
			handler = addonHandler.Create
		case "addons.updated":
			handler = addonHandler.Update
		case "addons.deleted":
			handler = addonHandler.Delete
		}
		if err := handler(parsedEvent); err != nil {
			h.Logger.Error(err, "Handler execution failed", "event", parsedEvent.Type)
		}
	})

	stream.Subscribe(subscription)
}

func (h *Handler) projectHandler() *ProjectHandler {
	return &ProjectHandler{
		Logger: h.Logger,
		Client: h.Client,
	}
}

func (h *Handler) addonHandler() *AddonHandler {
	return &AddonHandler{
		Logger: h.Logger,
		Client: h.Client,
	}
}
