package events

import (
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

func (h *Handler) RegisterSubscriptions(stream *action_cable.Stream) {
	subscription := action_cable.NewSubscription(map[string]string{
		"cluster_id": "1",
		"channel":    "ClusterChannel",
	})

	projectHandler := h.projectHandler()
	//addonHandler := h.addonHandler()

	subscription.Handler(func(event *action_cable.ActionCableEvent) {
		var handler action_cable.HandlerFunc
		switch event.Type {
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
		}
		handler(event)
	})

	//router.Subscribe(AddonCreatedEvent, addonHandler.Create)
	//router.Subscribe(AddonUpdatedEvent, addonHandler.Update)
	//router.Subscribe(AddonDeletedEvent, addonHandler.Delete)
}

func (h *Handler) ackMessage(eventId string) error {
	//return h.SendFunc(AckEvent{EventId: eventId})
	return nil
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
