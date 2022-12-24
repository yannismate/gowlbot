package logging

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
	"time"
)

func (m *Module) handleLoggingStatusCommand(interaction *discordgo.Interaction) {

	var settings []GuildLoggingSetting

	result := m.db.Where(&GuildLoggingSetting{GuildID: interaction.GuildID}).Find(&settings)

	if result.Error != nil {
		m.logger.Error("Error fetching guild logging settings", zap.Any("guild", interaction.GuildID), zap.Any("interaction", interaction.ID), zap.Error(result.Error))
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "There was an error fetching the logging settings for this server.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.Any("guild", interaction.GuildID), zap.Any("interaction", interaction.ID), zap.Error(err))
		}
		return
	}

	settingsMap := make(map[LogType]GuildLoggingSetting)
	for _, setting := range settings {
		settingsMap[setting.LogType] = setting
	}

	getEnabledString := func(logType LogType) string {
		setting, ok := settingsMap[logType]
		if ok && setting.Enabled {
			if len(setting.LoggingChannelID) > 0 {
				return "Enabled (<#" + setting.LoggingChannelID + ">)"
			} else {
				return "Enabled (Channel not configured!)"
			}
		} else {
			return "Disabled"
		}
	}

	err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Type:  discordgo.EmbedTypeRich,
					Title: "Current Logging Status",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  MessageEdit.ToReadableString(),
							Value: getEnabledString(MessageEdit),
						},
						{
							Name:  MessageDelete.ToReadableString(),
							Value: getEnabledString(MessageDelete),
						},
						{
							Name:  MemberJoin.ToReadableString(),
							Value: getEnabledString(MemberJoin),
						},
						{
							Name:  MemberLeave.ToReadableString(),
							Value: getEnabledString(MemberLeave),
						},
						{
							Name:  MemberRoleChange.ToReadableString(),
							Value: getEnabledString(MemberRoleChange),
						},
					},
					Color:     util.EmbedColorInfo,
					Timestamp: time.Now().Format(time.RFC3339),
					Footer: &discordgo.MessageEmbedFooter{
						Text: "gowlbot " + util.GetVersionString(),
					},
				},
			},
		},
	})
	if err != nil {
		m.logger.Error("Error responding to interaction", zap.Any("guild", interaction.GuildID), zap.Any("interaction", interaction.ID), zap.Error(err))
	}
}
