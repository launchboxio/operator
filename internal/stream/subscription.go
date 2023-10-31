package stream

type Subscription struct {
	Channel   string `json:"channel"`
	ClusterId int    `json:"cluster_id"`
}

type SubscriptionEvent struct {
	Channel   string `json:"channel"`
	ClusterId int    `json:"cluster_id"`
}

//
//func NewSubscription(stream *Stream, channel string, clusterId int) (*Subscription, error) {
//	sub := &Subscription{
//		Channel:   channel,
//		ClusterId: clusterId,
//	}
//
//	identifier, err := json.Marshal(&SubscriptionEvent{
//		Channel:   channel,
//		ClusterId: clusterId,
//	})
//	if err != nil {
//		return nil, err
//	}
//	sub.identifier = identifier
//	sub.Stream = stream
//	return sub, nil
//}
//
//func (s *Subscription) Notify(command string, event Event) error {
//	data, err := event.Marshal()
//	if err != nil {
//		return err
//	}
//	baseEvent := BaseEvent{
//		Command:    command,
//		Data:       data,
//		Identifier: s.identifier,
//	}
//	encoded, err := baseEvent.Marshal()
//	return s.Stream.Send(encoded)
//}
