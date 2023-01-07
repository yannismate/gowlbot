package notifications

import (
	"github.com/yannismate/gowlbot/internal/twitch"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	logger *zap.Logger
	db     *gorm.DB
	twitch *twitch.Twitch
}

func ProvideNotificationModule(logger *zap.Logger, db *gorm.DB, twitch *twitch.Twitch) *Module {
	return &Module{logger: logger, db: db, twitch: twitch}
}

func (m *Module) Name() string {
	return "notifications"
}

func (m *Module) Start() error {
	res, err := m.twitch.GetStreams([]string{"140551421"})
	if err != nil {
		return err
	}
	m.logger.Info("Got twitch user", zap.Any("user", res))
	return nil
}
