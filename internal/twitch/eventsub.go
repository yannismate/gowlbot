package twitch

import (
	"context"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const twitchEventsubUrl = "wss://eventsub-beta.wss.twitch.tv/ws"

type EventSub struct {
	ws                      *websocket.Conn
	stopWs                  *context.CancelFunc
	logger                  *zap.Logger
	listeners               map[string][]eventHandler
	recentMessageIDs        []string
	sessionID               string
	reconnectBackoffSeconds int
}

func ProvideEventSub(logger *zap.Logger) (*EventSub, error) {
	logger.Info("Connecting to Twitch EventSub", zap.String("url", twitchEventsubUrl))
	es := EventSub{logger: logger, listeners: make(map[string][]eventHandler), reconnectBackoffSeconds: 2}
	es.AddHandler(func(event *SessionWelcomeEvent) {
		es.sessionID = event.Session.ID
	})
	es.AddHandler(func(event *SessionReconnectEvent) {
		es.connectWebsocket(&event.Session.ReconnectURL)
	})

	err := es.connectWebsocket(nil)
	if err != nil {
		return nil, err
	}
	return &es, nil
}

func (es *EventSub) connectWebsocket(reconnectUrl *string) error {
	if es.ws != nil && es.stopWs != nil {
		(*es.stopWs)()
		es.ws.Close()
	}
	ctx, cancelCtx := context.WithCancel(context.Background())
	url := twitchEventsubUrl
	if reconnectUrl != nil {
		url = *reconnectUrl
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		cancelCtx()
		return err
	}
	go func() {
		defer conn.Close()
		defer cancelCtx()
		defer func() {
			for {
				es.logger.Warn("Twitch Eventsub disconnected, waiting " + strconv.Itoa(es.reconnectBackoffSeconds) + " seconds before reconnecting")
				time.Sleep(time.Second * time.Duration(es.reconnectBackoffSeconds))
				err := es.connectWebsocket(nil)
				if err != nil {
					es.logger.Error("Error reconnecting to Twitch EventSub", zap.Error(err))
					es.reconnectBackoffSeconds = es.reconnectBackoffSeconds * 2
				} else {
					break
				}
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					es.logger.Error("There was an error reading from the Twitch EventSub Socket", zap.Error(err))
					cancelCtx()
					break
				}
				if msgType == websocket.PingMessage {
					conn.WriteControl(websocket.PongMessage, msg, time.Now().Add(time.Duration(time.Second*2)))
					continue
				}
				parsedMessage, err := es.parseMessage(msg)
				if err != nil {
					es.logger.Warn("There was an Error parsing a Twitch EventSub message", zap.Error(err))
					continue
				}
				es.distributeEvent(parsedMessage)

			}
		}
	}()
	es.stopWs = &cancelCtx
	es.ws = conn
	return nil
}
