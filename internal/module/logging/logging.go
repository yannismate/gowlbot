package logging

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Module struct {
	discord *discordgo.Session
	logger  *zap.Logger
}

func ProvideLoggingModule(discord *discordgo.Session, logger *zap.Logger) Module {
	return Module{discord: discord, logger: logger}
}

func (m *Module) Start() {
	m.logger.Info("Starting Logging module.")
}
