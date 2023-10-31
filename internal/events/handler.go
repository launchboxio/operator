package events

import (
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/launchboxio/operator/internal/stream"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	Logger   logr.Logger
	Client   client.Client
	router   *EventRouter
	SendFunc func(event stream.Event) error
}

type SendFunc func(event stream.Event) error

func New(logger logr.Logger, client client.Client, sender SendFunc) *Handler {
	handler := &Handler{
		Logger:   logger,
		Client:   client,
		SendFunc: sender,
	}
	handler.registerSubscriptions()
	return handler
}

func (h *Handler) Listen(message []byte) error {
	h.Logger.Info(string(message))

	event, err := h.parse(message)
	if err != nil {
		return err
	}

	return h.processEvent(event)
}

func (h *Handler) registerSubscriptions() {
	router := NewRouter()
	router.Subscribe(PingEvent, func(event *ActionCableEvent) error {
		return nil
	})
	router.Subscribe("welcome", func(event *ActionCableEvent) error {
		return nil
	})
	router.Subscribe("confirm_subscription", func(event *ActionCableEvent) error {
		return nil
	})
	router.Subscribe("test", func(event *ActionCableEvent) error {
		return nil
	})

	projectHandler := h.projectHandler()
	router.Subscribe(ProjectCreatedEvent, projectHandler.Create)
	router.Subscribe(ProjectUpdatedEvent, projectHandler.Update)
	router.Subscribe(ProjectPausedEvent, projectHandler.Pause)
	router.Subscribe(ProjectResumedEvent, projectHandler.Resume)
	router.Subscribe(ProjectDeletedEvent, projectHandler.Delete)

	addonHandler := h.addonHandler()
	router.Subscribe(AddonCreatedEvent, addonHandler.Create)
	router.Subscribe(AddonUpdatedEvent, addonHandler.Update)
	router.Subscribe(AddonDeletedEvent, addonHandler.Delete)

	h.router = router
}

func (h *Handler) processEvent(event *ActionCableEvent) error {
	if err := h.ackMessage(event.Message.Id); err != nil {
		return err
	}
	if err := h.router.Dispatch(event); err != nil {
		return err
	}
	// TODO: Emit a response
	return nil
}

func (h *Handler) parse(message []byte) (*ActionCableEvent, error) {
	actionCableEvent := &ActionCableEvent{}
	var rawMessage map[string]interface{}
	if err := json.Unmarshal(message, &rawMessage); err != nil {
		return nil, err
	}
	messageType, ok := rawMessage["type"]
	if ok {
		if messageType == "ping" || messageType == "confirm_subscription" {
			return &ActionCableEvent{
				Message: ActionCableEventMessage{
					Type: messageType.(string),
				},
			}, nil
		}
	}

	if err := actionCableEvent.Unmarshal(message); err != nil {
		return nil, err
	}

	return actionCableEvent, nil
}

func (h *Handler) ackMessage(eventId string) error {
	return h.SendFunc(AckEvent{EventId: eventId})
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
