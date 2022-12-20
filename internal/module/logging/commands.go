package logging

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
)

const (
	CommandNameLogging       = "logging"
	CommandOptionStatus      = "status"
	CommandOptionUpdate      = "update"
	CommandOptionLoggingType = "logging_type"
	CommandOptionEnabledCmd  = "enabled"
	CommandOptionEnabled     = "enabled"
	CommandOptionFormatCmd   = "format"
	CommandOptionFormat      = "format"
	CommandOptionChannelCmd  = "channel"
	CommandOptionChannel     = "channel"
)

func (m *Module) registerSlashCommandListeners() {
	m.discord.AddHandler(m.handleInteractionCreation)
}

func (m *Module) handleInteractionCreation(_ *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := interaction.Data.(discordgo.ApplicationCommandInteractionData)

	if data.Name != CommandNameLogging {
		return
	}

	optionMap := util.ExtractOptionsMap(data.Options)

	if _, ok := optionMap[CommandOptionStatus]; ok {
		m.handleLoggingStatusCommand(interaction.Interaction)
	} else if _, ok = optionMap[CommandOptionUpdate]; ok {
		m.handleLoggingUpdateCommand(interaction.Interaction, optionMap)
	}
}

func (m *Module) MigrateSlashCommands(guildID string, oldCmds []*discordgo.ApplicationCommand) {
	var cmdDmPermission = false
	var adminMemberPermission int64 = discordgo.PermissionAdministrator
	var version = "logging-1.4"

	loggingTypeOption := discordgo.ApplicationCommandOption{
		Name:        CommandOptionLoggingType,
		Description: "Logging Type",
		Type:        discordgo.ApplicationCommandOptionString,
		Choices: []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  MessageEdit.ToReadableString(),
				Value: MessageEdit,
			},
			{
				Name:  MessageDelete.ToReadableString(),
				Value: MessageDelete,
			},
			{
				Name:  MemberJoin.ToReadableString(),
				Value: MemberJoin,
			},
			{
				Name:  MemberLeave.ToReadableString(),
				Value: MemberLeave,
			},
		},
		Required: true,
	}

	newLoggingCmd := discordgo.ApplicationCommand{
		Name:                     CommandNameLogging,
		Version:                  version,
		Description:              "View or update the bots logging settings for this server",
		DefaultMemberPermissions: &adminMemberPermission,
		DMPermission:             &cmdDmPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        CommandOptionStatus,
				Description: "View current logging settings",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        CommandOptionUpdate,
				Description: "Modify settings for a specific logging type",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        CommandOptionEnabledCmd,
						Description: "Enable or disable logging of this type",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							&loggingTypeOption,
							{
								Name:        CommandOptionEnabled,
								Description: "Enabled",
								Type:        discordgo.ApplicationCommandOptionBoolean,
								Required:    true,
							},
						},
					},
					{
						Name:        CommandOptionFormatCmd,
						Description: "Set logging format",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							&loggingTypeOption,
							{
								Name:        CommandOptionFormat,
								Description: "Format",
								Type:        discordgo.ApplicationCommandOptionString,
								Required:    true,
							},
						},
					},
					{
						Name:        CommandOptionChannelCmd,
						Description: "Set logging channel",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							&loggingTypeOption,
							{
								Name:        CommandOptionChannel,
								Description: "Channel",
								Type:        discordgo.ApplicationCommandOptionChannel,
								Required:    true,
							},
						},
					},
				},
			},
		},
	}

	commandHandled := false
	for _, oldCmd := range oldCmds {
		if oldCmd.Name == "logging" {
			if oldCmd.Version != newLoggingCmd.Version {
				_, err := m.discord.ApplicationCommandEdit(m.config.Discord.ApplicationID, guildID, oldCmd.ID, &newLoggingCmd)
				if err != nil {
					m.logger.Error("Error updating logging command", zap.Any("guild", guildID), zap.Any("command", oldCmd.ID), zap.Error(err))
					continue
				}
			}
			commandHandled = true
		}
	}
	if !commandHandled {
		_, err := m.discord.ApplicationCommandCreate(m.config.Discord.ApplicationID, guildID, &newLoggingCmd)
		if err != nil {
			m.logger.Error("Error creating logging command", zap.Any("guild", guildID), zap.Error(err))
		}
	}
}
