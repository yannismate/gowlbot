package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/config"
	"go.uber.org/zap"
)

func ProvideDiscordClient(cfg *config.OwlBotConfig, logger *zap.Logger) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.BotToken)
	session.Identify.Intents = discordgo.IntentsAll

	// Messages will be stored in redis
	session.State.MaxMessageCount = 0
	// necessary for Role Change Logging
	session.State.TrackMembers = true
	session.State.TrackEmojis = false
	session.State.TrackChannels = false
	session.State.TrackPresences = false
	session.State.TrackRoles = false
	session.State.TrackThreadMembers = false
	session.State.TrackThreads = false
	session.State.TrackVoice = false

	if err != nil {
		logger.Error("Error creating discord client", zap.Error(err))
		return nil, err
	}

	err = session.Open()
	if err != nil {
		logger.Error("Error opening client session", zap.Error(err))
		return nil, err
	}

	name := session.State.User.Username
	discriminator := session.State.User.Discriminator
	logger.Info(fmt.Sprintf("Connected to discord as user %s#%s", name, discriminator))

	return session, nil
}
