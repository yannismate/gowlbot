package logging

import (
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v9"
	"github.com/yannismate/gowlbot/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	config  *config.OwlBotConfig
	discord *discordgo.Session
	db      *gorm.DB
	cache   *redis.Client
	logger  *zap.Logger
}

func ProvideLoggingModule(config *config.OwlBotConfig, discord *discordgo.Session, db *gorm.DB, cache *redis.Client, logger *zap.Logger) Module {
	return Module{config: config, discord: discord, db: db, cache: cache, logger: logger}
}

func (m *Module) Start() {
	m.logger.Info("Starting Logging module.")
	err := m.db.AutoMigrate(&GuildLoggingSetting{})
	if err != nil {
		m.logger.Error("Could not prepare database for logging module", zap.Error(err))
	}
	m.registerMessageListeners()
	m.registerMemberJoinLeaveListeners()
	m.registerMemberRoleListeners()
	m.registerSlashCommandListeners()
}
