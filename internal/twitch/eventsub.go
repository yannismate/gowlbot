package twitch

import (
	"context"
	"github.com/yannismate/gowlbot/internal/config"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const twitchEventsubUrl = "wss://eventsub-beta.wss.twitch.tv/ws"

type EventSub struct {
	cfg                     *config.OwlBotConfig
	ws                      *websocket.Conn
	stopWs                  *context.CancelFunc
	logger                  *zap.Logger
	listeners               map[string][]eventHandler
	recentMessageIDs        []string
	sessionID               string
	reconnectBackoffSeconds int
	appAccessToken          string
	appAccessTokenExpiresAt time.Time
}

func ProvideEventSub(logger *zap.Logger, cfg *config.OwlBotConfig) (*EventSub, error) {
	es := EventSub{cfg: cfg, logger: logger, listeners: make(map[string][]eventHandler), reconnectBackoffSeconds: 2}
	es.AddHandler(func(event *SessionWelcomeEvent) {
		logger.Info("Connected to Twitch EventSub", zap.Any("session_id", event.Session.ID))
		es.sessionID = event.Session.ID
	})
	es.AddHandler(func(event *SessionReconnectEvent) {
		es.connectWebsocket(&event.Session.ReconnectURL)
	})

	logger.Info("Connecting to Twitch EventSub", zap.String("url", twitchEventsubUrl))
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
					err := conn.WriteControl(websocket.PongMessage, msg, time.Now().Add(time.Second*2))
					if err != nil {
						es.logger.Error("There was an error responding to a ping on the Twitch EventSub Socket", zap.Error(err))
						cancelCtx()
						return
					}
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
