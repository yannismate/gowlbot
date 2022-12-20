package module

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/config"
	"github.com/yannismate/gowlbot/internal/module/logging"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func GetRegisteredModules() []interface{} {
	var modules []interface{}

	modules = append(modules, logging.ProvideLoggingModule)

	return modules
}

type ModuleStarter struct {
	fx.In
	Logger  *zap.Logger
	Config  *config.OwlBotConfig
	Discord *discordgo.Session
	Logging logging.Module
}

func StartModules(starter ModuleStarter) {
	starter.Logger.Info("Starting Bot Modules")
	starter.Logging.Start()

	starter.Logger.Info("Migrating Guilds")
	starter.Discord.AddHandler(func(_ *discordgo.Session, guildJoin *discordgo.GuildCreate) {
		starter.migrateGuild(guildJoin.ID)
	})
	for _, guild := range starter.Discord.State.Guilds {
		starter.migrateGuild(guild.ID)
	}

	starter.Logger.Info("Module startup completed")
}

func (ms *ModuleStarter) migrateGuild(guildID string) {
	oldCmds, err := ms.Discord.ApplicationCommands(ms.Config.Discord.ApplicationID, guildID)
	if err != nil {
		ms.Logger.Error("Error fetching old application commands", zap.Any("guild", guildID), zap.Error(err))
		return
	}

	ms.Logging.MigrateSlashCommands(guildID, oldCmds)
}
