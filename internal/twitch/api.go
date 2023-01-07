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

const (
	twitchOauthTokenEndpoint   = "https://id.twitch.tv/oauth2/token"
	twitchHelixStreamsEndpoint = "https://api.twitch.tv/helix/streams"
	twitchHelixUsersEndpoint   = "https://api.twitch.tv/helix/users"
)

var (
	ErrStatusCodeFailed = errors.New("non 2xx status code")
	ErrParamListTooLong = errors.New("the parameter list is too long")
)

type clientCredentialsResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type StreamsResponse struct {
	Data []struct {
		ID           string `json:"id"`
		UserID       string `json:"user_id"`
		UserLogin    string `json:"user_login"`
		UserName     string `json:"user_name"`
		GameName     string `json:"game_name"`
		Type         string `json:"type"`
		Title        string `json:"title"`
		ThumbnailURL string `json:"thumbnail_url"`
	} `json:"data"`
	Pagination struct {
		Cursor *string `json:"cursor"`
	} `json:"pagination"`
}

type UsersResponse struct {
	Data []struct {
		ID          string `json:"id"`
		Login       string `json:"login"`
		DisplayName string `json:"display_name"`
	} `json:"data"`
}

func (t *Twitch) updateAppAccessToken() error {
	t.logger.Info("Fetching new Twitch app access token")
	params := url.Values{
		"client_id":     []string{t.cfg.Twitch.ClientID},
		"client_secret": []string{t.cfg.Twitch.ClientSecret},
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
	if res.StatusCode/100 != 2 {
		return ErrStatusCodeFailed
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
	t.appAccessToken = parsedRes.AccessToken
	t.appAccessTokenExpiresAt = time.Now().Add(time.Second * time.Duration(parsedRes.ExpiresIn))
	return nil
}

func (t *Twitch) GetUsers(userLogins []string) (*UsersResponse, error) {
	if len(userLogins) > 100 {
		return nil, ErrParamListTooLong
	}
	if t.appAccessTokenExpiresAt.Before(time.Now()) {
		err := t.updateAppAccessToken()
		if err != nil {
			return nil, err
		}
	}

	params := url.Values{"login": userLogins}

	req, err := http.NewRequest(http.MethodGet, twitchHelixUsersEndpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-Id", t.cfg.Twitch.ClientID)
	req.Header.Set("Authorization", "Bearer "+t.appAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 != 2 {
		return nil, ErrStatusCodeFailed
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	parsedRes := UsersResponse{}
	err = json.Unmarshal(resBody, &parsedRes)
	if err != nil {
		return nil, err
	}
	return &parsedRes, nil
}

func (t *Twitch) GetStreams(channelIds []string) (*StreamsResponse, error) {
	if len(channelIds) > 100 {
		return nil, ErrParamListTooLong
	}
	if t.appAccessTokenExpiresAt.Before(time.Now()) {
		err := t.updateAppAccessToken()
		if err != nil {
			return nil, err
		}
	}

	params := url.Values{
		"user_id": channelIds,
		"first":   []string{"100"},
	}

	req, err := http.NewRequest(http.MethodGet, twitchHelixStreamsEndpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-Id", t.cfg.Twitch.ClientID)
	req.Header.Set("Authorization", "Bearer "+t.appAccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		return nil, ErrStatusCodeFailed
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	parsedRes := StreamsResponse{}
	err = json.Unmarshal(resBody, &parsedRes)
	if err != nil {
		return nil, err
	}
	if parsedRes.Pagination.Cursor != nil {
		t.logger.Warn("Twitch returned pagination parameters when none were expected", zap.Any("channel_ids", channelIds))
	}
	return &parsedRes, nil
}
