package events

import (
	"errors"
)

type EventRouter struct {
	routes map[string]EventHandler
}

func NewRouter() *EventRouter {
	return &EventRouter{
		routes: map[string]EventHandler{},
	}
}

func (r *EventRouter) Subscribe(eventName string, handler EventHandler) {
	r.routes[eventName] = handler
}

func (r *EventRouter) Dispatch(event *ActionCableEvent) error {
	if handler, ok := r.routes[event.Message.Type]; ok {
		// TODO: We should ack the messages
		// handle the message
		return handler(event)
		// TODO: Maybe also support automatically emitting the response
	}
	return errors.New("Handler not registered for " + event.Message.Type)
}
