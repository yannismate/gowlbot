package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type subscriptionRequest struct {
	Type      SubscriptionType               `json:"type"`
	Version   string                         `json:"version"`
	Condition map[string]string              `json:"condition"`
	Transport subscriptionWebsocketTransport `json:"transport"`
}

type subscriptionWebsocketTransport struct {
	Method    string `json:"method"`
	SessionID string `json:"session_id"`
}

type clientCredentialsResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

const (
	twitchOauthTokenEndpoint        = "https://id.twitch.tv/oauth2/token"
	twitchHelixSubscriptionEndpoint = "https://api.twitch.tv/helix/eventsub/subscriptions"
)

var (
	ErrWebsocketNotConnected = errors.New("could not create a subscription since the websocket is not currently connected")
	ErrResponseCodeNoSuccess = errors.New("non 2xx response")
)

func (es *EventSub) UpdateAppAccessToken() error {
	es.logger.Info("Fetching new Twitch app access token")
	params := url.Values{
		"client_id":     []string{es.cfg.Twitch.ClientID},
		"client_secret": []string{es.cfg.Twitch.ClientSecret},
		"grant_type":    []string{"client_credentials"},
	}

	req, err := http.NewRequest("POST", twitchOauthTokenEndpoint, bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	parsedRes := clientCredentialsResponse{}
	err = json.Unmarshal(resBody, &parsedRes)
	if err != nil {
		return err
	}
	es.appAccessToken = parsedRes.AccessToken
	es.appAccessTokenExpiresAt = time.Now().Add(time.Second * time.Duration(parsedRes.ExpiresIn))
	return nil
}

func (es *EventSub) AddSubscription(subscriptionType SubscriptionType, conditions map[string]string) error {
	if len(es.sessionID) == 0 {
		return ErrWebsocketNotConnected
	}
	if len(es.appAccessToken) == 0 || es.appAccessTokenExpiresAt.Before(time.Now()) {
		err := es.UpdateAppAccessToken()
		if err != nil {
			return err
		}
	}
	requestData := subscriptionRequest{
		Type:      subscriptionType,
		Version:   "1",
		Condition: conditions,
		Transport: subscriptionWebsocketTransport{
			Method:    "websocket",
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
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Client-Id", es.cfg.Twitch.ClientID)
	request.Header.Set("Authorization", "Bearer "+es.appAccessToken)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode/100 != 2 {
		resBody, _ := ioutil.ReadAll(res.Body)
		es.logger.Info("Twitch API responded with error", zap.Any("body", string(resBody)), zap.Any("code", res.StatusCode))
		return ErrResponseCodeNoSuccess
	}
	return nil
}
