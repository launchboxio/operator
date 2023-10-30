package events

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Handler struct {
	Logger logr.Logger
	Client client.Client
	router *EventRouter
	conn   *websocket.Conn
	send   chan []byte
}

func NewHandler(conn *websocket.Conn, logger logr.Logger, client client.Client) *Handler {
	handler := &Handler{
		Logger: logger,
		Client: client,
		conn:   conn,
	}
	handler.registerSubscriptions()
	return handler
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

func (h *Handler) Listen(ctx context.Context, channel string, clusterId int) error {

	done := make(chan struct{})
	h.send = make(chan []byte)

	// Start our listener, which receives, processes, and routes events
	go func() {
		defer close(done)
		h.Logger.Info("Starting listener")
		h.listener()
	}()

	if err := h.subscribe(channel, clusterId); err != nil {
		h.Logger.Error(err, "Failed to subscribe to cluster channel")
		return err
	}

	for {
		select {
		case <-done:
			h.Logger.Info("Done received")
			return nil
		case msg := <-h.send:
			h.Logger.Info(string(msg))
			if err := h.sendMessage(msg); err != nil {
				h.Logger.Error(err, "Failed sending message")
			}
		case <-ctx.Done():
			h.Logger.Info("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := h.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

// sendMessage emits a message back over the websocket connection
func (h *Handler) sendMessage(message []byte) error {
	return h.conn.WriteMessage(websocket.TextMessage, message)
}

// listener receives messages on the socket stream,
// and queues them for processing
func (h *Handler) listener() {
	for {
		_, message, err := h.conn.ReadMessage()
		h.Logger.Info(string(message))
		if err != nil {
			h.Logger.Error(err, "Failed reading message")
			continue
		}
		event, err := h.parse(message)
		if err != nil {
			h.Logger.Error(err, "Failed parsing message")
			continue
		}
		if err := h.processEvent(event); err != nil {
			h.Logger.Error(err, "Failed processing event")
		}
	}
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
	ack := &AckResponse{EventId: eventId}
	ackBytes, err := json.Marshal(ack)
	if err != nil {
		return err
	}
	return h.sendMessage(ackBytes)
}

func (h *Handler) subscribe(channel string, clusterId int) error {
	identifier := &struct {
		Channel   string `json:"channel"`
		ClusterId int    `json:"cluster_id"`
	}{
		Channel:   channel,
		ClusterId: clusterId,
	}
	identBytes, err := json.Marshal(identifier)
	if err != nil {
		return err
	}
	subscription := &struct {
		Command    string `json:"command"`
		Identifier string `json:"identifier"`
	}{
		Command:    "subscribe",
		Identifier: string(identBytes),
	}
	sub, err := json.Marshal(subscription)
	if err != nil {
		return err
	}
	return h.sendMessage(sub)
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
