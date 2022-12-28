package twitch

import (
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestEventHandling(t *testing.T) {
	es := EventSub{listeners: make(map[string][]eventHandler)}
	wasCalled := false
	es.AddHandler(func (event *NotificationEvent) {
		wasCalled = true
	})
	es.AddHandler(func (event *SessionWelcomeEvent) {
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
	es := EventSub{listeners: map[string][]eventHandler{}, logger: logger}
	es.AddHandler(func (event *SessionWelcomeEvent) {
		logger.Info("welcome")
		wg.Done()
	})
	err := es.connectWebsocket(nil)
	if err != nil {
		t.Error(err)
	}
	wg.Wait()
}