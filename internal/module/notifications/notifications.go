package notifications

import (
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v9"
	"github.com/yannismate/gowlbot/internal/twitch"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	logger  *zap.Logger
	db      *gorm.DB
	twitch  *twitch.Twitch
	discord *discordgo.Session
	cache   *redis.Client
}

func ProvideNotificationModule(logger *zap.Logger, db *gorm.DB, twitch *twitch.Twitch, discord *discordgo.Session, cache *redis.Client) *Module {
	return &Module{logger: logger, db: db, twitch: twitch, discord: discord, cache: cache}
}

func (m *Module) Name() string {
	return "notifications"
}

func (m *Module) Start() error {
	err := m.db.AutoMigrate(&GuildNotification{})
	if err != nil {
		return err
	}

	m.registerSlashCommandListeners()
	m.startTwitchUpdateTimer()

	return nil
}
