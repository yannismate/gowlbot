package twitch

type eventHandler interface {
	Type() string
	Handle(interface{})
}

type eventInstance interface {
	Type() string
	Instance() interface{}
}

func (es *EventSub) distributeEvent(parsedMessage eventInstance) {
	handlers, ok := es.listeners[parsedMessage.Type()]
	if !ok {
		return
	}
	for _, handler := range handlers {
		go handler.Handle(parsedMessage.Instance())
	}
}

func (es *EventSub) AddHandler(handler interface{}) {
	newHandler := handlerForInterface(handler)
	if newHandler == nil {
		es.logger.Error("EventSub Handler not added, invalid interface type.")
		return
	}
	handlers, ok := es.listeners[newHandler.Type()]
	if !ok {
		handlers = make([]eventHandler, 0)
	}
	handlers = append(handlers, newHandler)
	es.listeners[newHandler.Type()] = handlers
}

const (
	SessionWelcomeEventType   = "session_welcome_event"
	SessionKeepaliveEventType = "session_keepalive_event"
	SessionReconnectEventType = "session_reconnect_event"
	NotificationEventType     = "notification_event"
)

type sessionWelcomeEventHandler func(*SessionWelcomeEvent)

func (eh sessionWelcomeEventHandler) Type() string {
	return SessionWelcomeEventType
}
func (eh sessionWelcomeEventHandler) Handle(i interface{}) {
	if t, ok := i.(*SessionWelcomeEvent); ok {
		eh(t)
	}
}

type sessionKeepaliveEventHandler func(*SessionKeepaliveEvent)

func (eh sessionKeepaliveEventHandler) Type() string {
	return SessionKeepaliveEventType
}
func (eh sessionKeepaliveEventHandler) Handle(i interface{}) {
	if t, ok := i.(*SessionKeepaliveEvent); ok {
		eh(t)
	}
}

type sessionReconnectEventHandler func(*SessionReconnectEvent)

func (eh sessionReconnectEventHandler) Type() string {
	return SessionReconnectEventType
}
func (eh sessionReconnectEventHandler) Handle(i interface{}) {
	if t, ok := i.(*SessionReconnectEvent); ok {
		eh(t)
	}
}

type notificationEventHandler func(*NotificationEvent)

func (eh notificationEventHandler) Type() string {
	return NotificationEventType
}
func (eh notificationEventHandler) Handle(i interface{}) {
	if t, ok := i.(*NotificationEvent); ok {
		eh(t)
	}
}

func handlerForInterface(itf interface{}) eventHandler {
	switch v := itf.(type) {
	case func(*SessionWelcomeEvent):
		return sessionWelcomeEventHandler(v)
	case func(*SessionKeepaliveEvent):
		return sessionKeepaliveEventHandler(v)
	case func(event *SessionReconnectEvent):
		return sessionReconnectEventHandler(v)
	case func(*NotificationEvent):
		return notificationEventHandler(v)
	}
	return nil
}
