package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type subscriptionRequest struct {
	Type      SubscriptionType  `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	Transport subscriptionWebsocketTransport `json:"transport"`
}

type subscriptionWebsocketTransport struct {
	Method    string `json:"method"`
	SessionID string `json:"session_id"`
}

const (
	twitchHelixSubscriptionEndpoint = "https://api.twitch.tv/helix/eventsub/subscriptions"
)

var (
	ErrWebsocketNotConnected = errors.New("could not create a subscription since the websocket is not currently connected")
	ErrResponseCodeNoSuccess = errors.New("non 2xx response")
)

func (es *EventSub) AddSubscription(subscriptionType SubscriptionType, conditions map[string]string) error {
	if len(es.sessionID) == 0 {
		return ErrWebsocketNotConnected
	}
	requestData := subscriptionRequest{
		Type: subscriptionType,
		Version: "1",
		Condition: conditions,
		Transport: subscriptionWebsocketTransport{
			Method: "websocket",
			SessionID: es.sessionID,
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", twitchHelixSubscriptionEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode / 100 != 2 {
		return ErrResponseCodeNoSuccess
	}
	return nil
}
