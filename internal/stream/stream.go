package stream

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
	"time"
)

type Stream struct {
	// Configuration
	Url       string
	Auth      clientcredentials.Config
	Channel   string
	ClusterId int

	Logger logr.Logger

	UseBacklog bool

	// Websocket connection resources
	conn        *websocket.Conn
	send        chan []byte
	recv        chan []byte
	isConnected bool

	// When disconnected, store a backlog of events
	// that we can later propagate on reconnection
	backlog [][]byte

	listeners []Listener
}

type Listener func(message []byte) error

func New(url string, auth clientcredentials.Config, channel string, clusterId int) *Stream {
	return &Stream{
		Url:        url,
		Auth:       auth,
		Channel:    channel,
		ClusterId:  clusterId,
		UseBacklog: false,
	}
}

func (s *Stream) Listen(ctx context.Context) error {
	token, err := s.Auth.Token(context.TODO())
	if err != nil {
		return err
	}
	u, err := url.Parse(s.Url)
	if err != nil {
		return err
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{"Bearer " + token.AccessToken},
	})
	if err != nil {
		return err
	}

	s.conn = c
	done := make(chan struct{})
	s.send = make(chan []byte)

	// Start our listener
	go func() {
		defer close(done)
		for {
			_, message, err := s.conn.ReadMessage()
			fmt.Println(string(message))
			if err != nil {
				s.Logger.Error(err, "Failed reading message")
				continue
			}
			for _, listener := range s.listeners {
				if err = listener(message); err != nil {
					s.Logger.Error(err, "Failed emitting to listener")
				}
			}
		}
	}()

	// Subscribe to the requested channel
	if err := s.subscribe(); err != nil {
		return err
	}
	s.isConnected = true

	// TODO: Process the backlog and clear it

	for {
		select {
		case <-done:
			s.Logger.Info("Done received")
			return nil
		case msg := <-s.send:
			fmt.Println(string(msg))
			if err := s.Send(msg); err != nil {
				s.Logger.Error(err, "Failed sending message")
			}
		case <-ctx.Done():
			s.Logger.Info("interrupt")
			err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

// Send puts data directly into the send channel, if connected.
// Otherwise queue into the backlog, so it
// can be streamed on reconnection
func (s *Stream) Send(data []byte) error {
	if s.isConnected {
		s.send <- data
		return nil
	}

	if s.UseBacklog {
		s.backlog = append(s.backlog, data)
	}

	return nil
}

func (s *Stream) AddListener(listener Listener) {
	s.listeners = append(s.listeners, listener)
}

func (s *Stream) subscribe() error {
	fmt.Println("Subscribing")
	subscriptionEvent := SubscriptionEvent{
		ClusterId: s.ClusterId,
		Channel:   s.Channel,
	}
	fmt.Println(subscriptionEvent)
	return s.Notify(subscriptionEvent)
}

// Notify is a helper for converting a base event object
// into a message we can transmit to HQ
func (s *Stream) Notify(event Event) error {
	data, err := event.Marshal()
	if err != nil {
		return err
	}
	return s.Send(data)
}
