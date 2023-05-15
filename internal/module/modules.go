package module

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/config"
	"github.com/yannismate/gowlbot/internal/discord"
	"github.com/yannismate/gowlbot/internal/module/logging"
	"github.com/yannismate/gowlbot/internal/module/notifications"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type BotModule interface {
	Name() string
	Start() error
	GetSlashCommands() []discord.VersionedSlashCommand
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

	migrateGuild := func(guildID string) error {
		oldCmds, err := smi.Discord.ApplicationCommands(smi.Config.Discord.ApplicationID, guildID)
		if err != nil {
			smi.Logger.Error("Error fetching old application commands", zap.Any("guild", guildID), zap.Error(err))
			return nil
		}

		for _, module := range moduleList {
			commands := module.GetSlashCommands()
			for _, newCmd := range commands {
				commandHandled := false
				for _, oldCmd := range oldCmds {
					if oldCmd.Name == newCmd.CmdName {
						if oldCmd.Version != newCmd.Version {
							_, err := smi.Discord.ApplicationCommandEdit(smi.Config.Discord.ApplicationID, guildID, oldCmd.ID, &newCmd.Command)
							if err != nil {
								smi.Logger.Error("Error updating command", zap.Any("guild", guildID), zap.Any("command", newCmd.CmdName), zap.Any("command_id", oldCmd.ID), zap.Error(err))
								return err
							}
						}
						commandHandled = true
					}
				}
				if !commandHandled {
					_, err := smi.Discord.ApplicationCommandCreate(smi.Config.Discord.ApplicationID, guildID, &newCmd.Command)
					if err != nil {
						smi.Logger.Error("Error creating command", zap.Any("guild", guildID), zap.Any("command", newCmd.CmdName), zap.Error(err))
						return err
					}
				}
			}
			if err != nil {
				smi.Logger.Error("Error migrating application commands", zap.String("guild", guildID), zap.String("module", module.Name()), zap.Error(err))
				continue
			}
		}
		return nil
	}

	smi.Logger.Info("Migrating Guilds")
	for _, guild := range smi.Discord.State.Guilds {
		err := migrateGuild(guild.ID)
		if err != nil {
			smi.Logger.Error("Error migrating guild application commands", zap.String("guild", guild.ID), zap.Error(err))
		}
	}
	smi.Discord.AddHandler(func(_ *discordgo.Session, guildJoin *discordgo.GuildCreate) {
		err := migrateGuild(guildJoin.ID)
		if err != nil {
			smi.Logger.Error("Error migrating guild application commands", zap.String("guild", guildJoin.ID), zap.Error(err))
		}
	})

	smi.Logger.Info("Module startup completed")
	return nil
}
