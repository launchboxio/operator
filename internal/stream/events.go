package stream

import "encoding/json"

type Event interface {
	Marshal() ([]byte, error)
}

type MessageEvent struct {
}

type BaseEvent struct {
	Command    string `json:"command"`
	Data       string `json:"data,omitempty"`
	Identifier string `json:"identifier"`
}

func (be BaseEvent) Marshal() ([]byte, error) {
	return json.Marshal(be)
}

type Builder struct {
	Command string
	Data    string
}

func (b *Builder) SetCommand(command string) *Builder {
	b.Command = command
	return b
}

func (b *Builder) SetEvent(event Event) *Builder {
	data, _ := event.Marshal()
	b.Data = string(data)
	return b
}
