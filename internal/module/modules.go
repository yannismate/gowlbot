package module

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/config"
	"github.com/yannismate/gowlbot/internal/module/logging"
	"github.com/yannismate/gowlbot/internal/module/notifications"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type BotModule interface {
	Name() string
	Start() error
	MigrateSlashCommands(guildID string, oldCommands []*discordgo.ApplicationCommand) error
}

func GetRegisteredModules() []interface{} {
	var modules []interface{}

	modules = append(modules, logging.ProvideLoggingModule)
	modules = append(modules, notifications.ProvideNotificationModule)

	return modules
}

type StartModuleInjection struct {
	fx.In
	Logger        *zap.Logger
	Config        *config.OwlBotConfig
	Discord       *discordgo.Session
	Logging       *logging.Module
	Notifications *notifications.Module
}

func StartModules(smi StartModuleInjection) error {
	moduleList := []BotModule{
		smi.Logging,
		smi.Notifications,
	}

	smi.Logger.Info("Starting Bot Modules")
	for _, module := range moduleList {
		smi.Logger.Info("Starting module '" + module.Name() + "'")
		err := module.Start()
		if err != nil {
			smi.Logger.Error("Error during module startup", zap.String("module", module.Name()), zap.Error(err))
			return err
		}
	}

	migrateGuild := func(guildID string) {
		oldCmds, err := smi.Discord.ApplicationCommands(smi.Config.Discord.ApplicationID, guildID)
		if err != nil {
			smi.Logger.Error("Error fetching old application commands", zap.Any("guild", guildID), zap.Error(err))
			return
		}

		for _, module := range moduleList {
			err = module.MigrateSlashCommands(guildID, oldCmds)
			if err != nil {
				smi.Logger.Error("Error migrating application commands", zap.String("guild", guildID), zap.String("module", module.Name()), zap.Error(err))
				continue
			}
		}
	}

	smi.Logger.Info("Migrating Guilds")
	for _, guild := range smi.Discord.State.Guilds {
		migrateGuild(guild.ID)
	}
	smi.Discord.AddHandler(func(_ *discordgo.Session, guildJoin *discordgo.GuildCreate) {
		migrateGuild(guildJoin.ID)
	})

	smi.Logger.Info("Module startup completed")
	return nil
}
