package twitch

type eventHandlerInstance struct {
	handler eventHandler
}

type eventHandler interface {
	Type() string
	Handle(interface{})
}

type eventInstance interface {
	Type() string
	Instance() interface{}
}

func (es *EventSub) distributeNotification(parsedMessage eventInstance) {
	handlers, ok := es.listeners[parsedMessage.Type()]
	if !ok {
		return
	}
	for _, handler := range handlers {
		handler.Handle(parsedMessage.Instance())
	}
}

func (es *EventSub) AddHandler(handler interface{}) {
	newHandler := handlerForInterface(handler)
	handlers, ok := es.listeners[newHandler.Type()]
	if !ok {
		handlers = make([]eventHandler, 0)
	}
	handlers = append(handlers, newHandler)
	es.listeners[newHandler.Type()] = handlers
}

const (
	MessageStreamOnlineEventType = "message_stream_online_event"
	MessageStreamOfflineEventType = "message_stream_offline_event"
)

type messageStreamOnlineEventHandler func(*MessageStreamOnlineEvent)
func (eh messageStreamOnlineEventHandler) Type() string {
	return MessageStreamOnlineEventType
}
func (eh messageStreamOnlineEventHandler) Handle(i interface{}) {
	if t, ok := i.(*MessageStreamOnlineEvent); ok {
		eh(t)
	}
}

type messageStreamOfflineEventHandler func(*MessageStreamOfflineEvent)
func (eh messageStreamOfflineEventHandler) Type() string {
	return MessageStreamOfflineEventType
}
func (eh messageStreamOfflineEventHandler) Handle(i interface{}) {
	if t, ok := i.(*MessageStreamOfflineEvent); ok {
		eh(t)
	}
}

func handlerForInterface(itf interface{}) eventHandler {
	switch v := itf.(type) {
	case func(*MessageStreamOnlineEvent):
		return messageStreamOnlineEventHandler(v)
	case func(*MessageStreamOfflineEvent):
		return messageStreamOfflineEventHandler(v)
	}
	return nil
}