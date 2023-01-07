package twitch

import (
	"github.com/yannismate/gowlbot/internal/config"
	"go.uber.org/zap"
	"time"
)

type Twitch struct {
	logger                  *zap.Logger
	cfg                     *config.OwlBotConfig
	appAccessToken          string
	appAccessTokenExpiresAt time.Time
}

func ProvideTwitch(logger *zap.Logger, cfg *config.OwlBotConfig) (*Twitch, error) {
	twitch := Twitch{
		logger: logger,
		cfg:    cfg,
	}
	err := twitch.updateAppAccessToken()
	if err != nil {
		return nil, err
	}

	return &twitch, nil
}
