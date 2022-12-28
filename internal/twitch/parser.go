package twitch

import (
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
)

var (
	ErrUnknownMessageType         = errors.New("unknown message type")
	ErrMessagePayloadNotParseable = errors.New("message payload could not be casted")
	ErrEventTooOld                = errors.New("the event was older than 10 minutes and should be ignored")
	ErrDuplicatedEvent            = errors.New("this event was previously received")
)

func (es *EventSub) parseMessage(msg []byte) (eventInstance, error) {
	event := eventSubEvent{}
	err := json.Unmarshal(msg, &event)
	if err != nil {
		return nil, err
	}
	if event.Metadata.MessageTimestamp.Before(time.Now().Add(time.Minute * 10)) {
		return nil, ErrEventTooOld
	}
	for _, recentID := range es.recentMessageIDs {
		if recentID == event.Metadata.MessageID {
			return nil, ErrDuplicatedEvent
		}
	}
	es.recentMessageIDs = append(es.recentMessageIDs, event.Metadata.MessageID)
	if len(es.recentMessageIDs) > 50 {
		es.recentMessageIDs = es.recentMessageIDs[1:]
	}

	es.logger.Info("Parsing message", zap.Any("msg", event), zap.Any("payload", event.Payload))
	payloadJson, _ := json.Marshal(event.Payload)

	var payloadIf eventInstance

	switch event.Metadata.MessageType {
	case MessageSessionWelcome:
		payloadIf = &SessionWelcomeEvent{}
	case MessageSessionKeepalive:
		payloadIf = &SessionKeepaliveEvent{}
	case MessageNotification:
		payloadIf = &NotificationEvent{}
	default:
		es.logger.Info("Unimplemented Twitch EventSub message", zap.Any("event", event))
		return nil, ErrUnknownMessageType
	}

	err = json.Unmarshal(payloadJson, payloadIf)
	if err != nil {
		return nil, err
	}
	return payloadIf, nil
}

type MessageType string

const (
	MessageSessionWelcome   MessageType = "session_welcome"
	MessageSessionKeepalive MessageType = "session_keepalive"
	MessageSessionReconnect MessageType = "session_reconnect"
	MessageNotification     MessageType = "notification"
)

type SubscriptionType string

const (
	SubscriptionStreamOnline  SubscriptionType = "stream.online"
	SubscriptionStreamOffline SubscriptionType = "stream.offline"
)

type eventSubEvent struct {
	Metadata struct {
		MessageID           string           `json:"message_id"`
		MessageType         MessageType      `json:"message_type"`
		MessageTimestamp    time.Time        `json:"message_timestamp"`
		SubscriptionType    SubscriptionType `json:"subscription_type"`
		SubscriptionVersion string           `json:"subscription_version"`
	} `json:"metadata"`
	Payload interface{} `json:"payload"`
}

type SessionWelcomeEvent struct {
	Session struct {
		ConnectedAt             time.Time `json:"connected_at"`
		ID                      string    `json:"id"`
		KeepaliveTimeoutSeconds int       `json:"keepalive_timeout_seconds"`
		ReconnectURL            string    `json:"reconnect_url"`
		Status                  string    `json:"status"`
	}
}

func (e *SessionWelcomeEvent) Type() string {
	return SessionWelcomeEventType
}
func (e *SessionWelcomeEvent) Instance() interface{} {
	return e
}

type SessionKeepaliveEvent struct {
}

func (e *SessionKeepaliveEvent) Type() string {
	return SessionKeepaliveEventType
}
func (e *SessionKeepaliveEvent) Instance() interface{} {
	return e
}

type SessionReconnectEvent struct {
	Session struct {
		ID                      string    `json:"id"`
		Status                  string    `json:"status"`
		KeepaliveTimeoutSeconds int       `json:"keepalive_timeout_seconds"`
		ReconnectURL            string    `json:"reconnect_url"`
		ConnectedAt             time.Time `json:"connected_at"`
	} `json:"session"`
}

func (e *SessionReconnectEvent) Type() string {
	return SessionReconnectEventType
}
func (e *SessionReconnectEvent) Instance() interface{} {
	return e
}

type NotificationEvent struct {
	Subscription struct {
		ID        string            `json:"id"`
		Status    string            `json:"status"`
		Type      SubscriptionType  `json:"type"`
		Version   string            `json:"version"`
		Cost      string            `json:"cost"`
		Condition map[string]string `json:"condition"`
		Transport struct {
			Method    string `json:"method"`
			SessionID string `json:"session_id"`
		} `json:"subscription"`
		CreatedAt time.Time `json:"created_at"`
	}
}

func (e *NotificationEvent) Type() string {
	return NotificationEventType
}
func (e *NotificationEvent) Instance() interface{} {
	return e
}
