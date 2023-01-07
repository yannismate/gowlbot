package twitch

import (
	"github.com/yannismate/gowlbot/internal/config"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func TestEventHandling(t *testing.T) {
	es := EventSub{listeners: make(map[string][]eventHandler)}
	wasCalled := false
	es.AddHandler(func(event *NotificationEvent) {
		wasCalled = true
	})
	es.AddHandler(func(event *SessionWelcomeEvent) {
		t.Error("Wrong event listener was called")
	})
	es.distributeEvent(&NotificationEvent{})
	if !wasCalled {
		t.Error("Correct event listener was not called")
	}
}

func TestEventSub(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	logger, _ := zap.NewDevelopment()
	es, _ := ProvideEventSub(logger, &config.OwlBotConfig{Twitch: config.TwitchConfig{ClientID: "XXXX", ClientSecret: "XXXX"}})
	time.Sleep(time.Second)
	err := es.AddSubscription(SubscriptionStreamOnline, map[string]string{"broadcaster_user_id": "140551421"})
	if err != nil {
		t.Error(err)
	}
}
