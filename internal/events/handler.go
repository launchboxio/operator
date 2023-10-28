package events

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"log"
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
	return &Handler{
		Logger: logger,
		Client: client,
		conn:   conn,
	}
}

func (h *Handler) registerSubscriptions() {
	router := &EventRouter{}

	projectHandler := h.projectHandler()
	router.Subscribe(ProjectCreatedEvent, projectHandler.Create)
	router.Subscribe(ProjectUpdatedEvent, projectHandler.Update)
	router.Subscribe(ProjectPausedEvent, projectHandler.Pause)
	router.Subscribe(ProjectResumedEvent, projectHandler.Resume)
	router.Subscribe(ProjectDeletedEvent, projectHandler.Delete)

	h.router = router
}

func (h *Handler) Listen(ctx context.Context) error {

	done := make(chan struct{})
	h.send = make(chan []byte)

	// Start our listener, which receives, processes, and routes events
	go func() {
		defer close(done)
		err := h.listener()
		if err != nil {
			h.Logger.Error(err, "Failed listening to socket messages")
		}
	}()

	for {
		select {
		case <-done:
			return nil
		case msg := <-h.send:
			if err := h.sendMessage(msg); err != nil {
				h.Logger.Error(err, "Failed sending message")
			}
		case <-ctx.Done():
			log.Println("interrupt")

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
func (h *Handler) listener() error {
	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			return err
		}
		event, err := h.parse(message)
		if err != nil {
			return err
		}
		if err := h.processEvent(event); err != nil {
			return err
		}
	}
}

func (h *Handler) processEvent(event Event) error {
	if err := h.ackMessage(event.Id); err != nil {
		return err
	}
	if err := h.router.Dispatch(event); err != nil {
		return err
	}
	// TODO: Emit a response
	return nil
}

func (h *Handler) parse(message []byte) (Event, error) {
	event := Event{}
	err := json.Unmarshal(message, &event)
	return event, err
}

func (h *Handler) ackMessage(eventId string) error {
	ack := &AckResponse{EventId: eventId}
	ackBytes, err := json.Marshal(ack)
	if err != nil {
		return err
	}
	return h.sendMessage(ackBytes)
}

func (h *Handler) projectHandler() *ProjectHandler {
	return &ProjectHandler{
		Logger: h.Logger,
		Client: h.Client,
	}
}
