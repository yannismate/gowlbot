package notifications

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/discord"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
)

const (
	CommandNameNotifications   = "notifications"
	CommandOptionListCmd       = "list"
	CommandOptionAddCmd        = "add"
	CommandOptionAddTwitchCmd  = "twitch"
	CommandOptionDeleteCmd     = "delete"
	CommandOptionDeleteId      = "id"
	CommandOptionChannel       = "channel"
	CommandOptionTwitchChannel = "twitch_channel"
)

func (m *Module) registerSlashCommandListeners() {
	m.discord.AddHandler(m.handleInteractionCreation)
}

func (m *Module) handleInteractionCreation(_ *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := interaction.Data.(discordgo.ApplicationCommandInteractionData)

	m.logger.Debug("hmmm", zap.Any("data", data))
	if data.Name != CommandNameNotifications {
		return
	}

	optionMap := util.ExtractOptionsMap(data.Options)
	m.logger.Debug("hmmm", zap.Any("optionMap", optionMap))

	if _, ok := optionMap[CommandOptionListCmd]; ok {
		m.handleNotificationsListCommand(interaction.Interaction)
	} else if _, ok = optionMap[CommandOptionAddCmd]; ok {
		m.handleNotificationsAddCommand(interaction.Interaction, optionMap)
	} else if _, ok = optionMap[CommandOptionDeleteCmd]; ok {
		m.handleNotificationsDeleteCommand(interaction.Interaction, optionMap)
	}
}

func (m *Module) GetSlashCommands() []discord.VersionedSlashCommand {
	var cmdDmPermission = false
	var adminMemberPermission int64 = discordgo.PermissionAdministrator
	var version = "notifications-1.0"

	newNotificationsCmd := discordgo.ApplicationCommand{
		Name:                     CommandNameNotifications,
		Version:                  version,
		Description:              "List or update the configured notifications for this server",
		DefaultMemberPermissions: &adminMemberPermission,
		DMPermission:             &cmdDmPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        CommandOptionListCmd,
				Description: "List currently configured notifications",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        CommandOptionAddCmd,
				Description: "Add a notification to this server",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        CommandOptionAddTwitchCmd,
						Description: "Twitch",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        CommandOptionChannel,
								Description: "Channel to send notifications to",
								Type:        discordgo.ApplicationCommandOptionChannel,
								Required:    true,
							},
							{
								Name:        CommandOptionTwitchChannel,
								Description: "Twitch channel name",
								Type:        discordgo.ApplicationCommandOptionString,
								Required:    true,
							},
						},
					},
				},
			},
			{
				Name:        CommandOptionDeleteCmd,
				Description: "Remove a configured notification from this server",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        CommandOptionDeleteId,
						Description: "Notification ID",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	}

	return []discord.VersionedSlashCommand{
		{
			Command: newNotificationsCmd,
			CmdName: CommandNameNotifications,
			Version: version,
		},
	}
}
