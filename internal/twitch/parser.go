package twitch

type MessageType string

const (
	MessageSessionWelcome   MessageType = "session_welcome"
	MessageSessionKeepalive MessageType = "session_keepalive"
	MessageNotification     MessageType = "notification"
)

func (es *EventSub) parseMessage(msg []byte) (*eventInstance, error) {
	// todo
	return nil, nil
}

type SessionWelcomeEvent struct {

}
func (e *SessionWelcomeEvent) Type() string {
	return "internal"
}
func (e *SessionWelcomeEvent) Instance() interface{} {
	return e
}

type SessionKeepaliveEvent struct {

}
func (e *SessionKeepaliveEvent) Type() string {
	return "internal"
}
func (e *SessionKeepaliveEvent) Instance() interface{} {
	return e
}

type MessageStreamOnlineEvent struct {

}
func (e *MessageStreamOnlineEvent) Type() string {
	return MessageStreamOnlineEventType
}
func (e *MessageStreamOnlineEvent) Instance() interface{} {
	return e
}

type MessageStreamOfflineEvent struct {

}
func (e *MessageStreamOfflineEvent) Type() string {
	return MessageStreamOfflineEventType
}
func (e *MessageStreamOfflineEvent) Instance() interface{} {
	return e
}