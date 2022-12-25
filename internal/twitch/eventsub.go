package twitch

import (
	"context"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const twitchEventsubUrl = "wss://eventsub-beta.wss.twitch.tv/ws"

type EventSub struct {
	ws     *websocket.Conn
	stopWs *context.CancelFunc
	logger *zap.Logger
	listeners map[string][]eventHandler
}

func ProvideEventSub(logger *zap.Logger) (*EventSub, error) {
	logger.Info("Connecting to Twitch EventSub", zap.String("url", twitchEventsubUrl))
	es := EventSub{logger: logger, listeners: make(map[string][]eventHandler)}
	err := es.connectWebsocket()
	if err != nil {
		return nil, err
	}
	return &es, nil
}

func (es *EventSub) connectWebsocket() error {
	if es.ws != nil && es.stopWs != nil {
		(*es.stopWs)()
		es.ws.Close()
	}
	ctx, cancelCtx := context.WithCancel(context.Background())
	conn, _, err := websocket.DefaultDialer.Dial(twitchEventsubUrl, nil)
	if err != nil {
		cancelCtx()
		return err
	}
	go func ()  {
		defer conn.Close()
		defer cancelCtx()
		for {
			select {
			case <- ctx.Done():
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					es.logger.Error("There was an error reading from the Twitch EventSub Socket", zap.Error(err))
					break
				}
				parsedMessage, err := es.parseMessage(msg)
				if err != nil {
					es.logger.Warn("There was an Error parsing a Twitch EventSub message", zap.Error(err))
					continue
				}
				es.distributeNotification(*parsedMessage)
				
			}
		}
		// todo: handle reconnect with backoff
	}()
	// todo: keepalive
	es.stopWs = &cancelCtx
	es.ws = conn
	return nil
}