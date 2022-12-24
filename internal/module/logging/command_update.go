package logging

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"time"
)

var (
	defaultLoggingFormats = map[LogType]string{
		MessageEdit:      "✏ <t:{time}> <#{channel_id}> **{author_full_name}** edited their message. Previous content: {previous_content}",
		MessageDelete:    "🗑 <t:{time}> <#{channel_id}> Message by **{author_full_name}** was deleted. Content: {previous_content}",
		MemberJoin:       "📥 <t:{time}> <@{member_id}> ({member_full_name}) joined the server. Total members: {guild_member_count}",
		MemberLeave:      "📤 <t:{time}> <@{member_id}> ({member_full_name}) left the server or got kicked. Total members: {guild_member_count}",
		MemberRoleChange: "👥 <t:{time}> **{member_full_name}**'s roles changed: `{role_changes}`",
	}
)

func (m *Module) handleLoggingUpdateCommand(interaction *discordgo.Interaction, optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	handleParseError := func(details string) {
		m.logger.Error("Error parsing logging update", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.String("details", details))
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "There was an error parsing your command inputs.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
		}
	}

	logTypeOption, ok := optionMap[CommandOptionLoggingType]
	if !ok || logTypeOption.Type != discordgo.ApplicationCommandOptionString {
		handleParseError("Logging type missing or wrong type")
		return
	}
	logType, ok := ParseLogType(logTypeOption.Value.(string))
	if !ok {
		handleParseError("Unknown logging type")
		return
	}

	settings := GuildLoggingSetting{}

	dbResult := m.db.Where(&GuildLoggingSetting{GuildID: interaction.GuildID, LogType: logType}).First(&settings)

	if dbResult.Error != nil && !errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		m.logger.Error("Error fetching logging settings from db", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID))
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "An internal error occurred.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
		}
		return
	}

	if enabledOption, ok := optionMap[CommandOptionEnabled]; ok {
		enabled := enabledOption.Value.(bool)
		if !ok {
			handleParseError("Enabled type not bool")
			return
		}

		settings.Enabled = enabled
	}

	if formatOption, ok := optionMap[CommandOptionFormat]; ok {
		format, ok := formatOption.Value.(string)
		if !ok {
			handleParseError("Format type not string")
			return
		}

		settings.Format = format
	}

	if channelOption, ok := optionMap[CommandOptionChannel]; ok {
		if channelOption.Type != discordgo.ApplicationCommandOptionChannel {
			handleParseError("Channel argument not channel")
			return
		}

		channel := channelOption.ChannelValue(m.discord)
		if channel.GuildID != interaction.GuildID {
			handleParseError("Channel is not in the same server")
			return
		}
		settings.LoggingChannelID = channel.ID
	}

	if dbResult.RowsAffected == 0 {
		settings.Format = defaultLoggingFormats[logType]
		settings.GuildID = interaction.GuildID
		settings.LogType = logType
		dbResult = m.db.Create(&settings)
	} else {
		dbResult = m.db.Save(&settings)
	}

	if dbResult.Error != nil {
		m.logger.Error("Error updating logging settings in db", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID))
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "An internal error occurred.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
		}
		return
	}

	loggingChannelStatus := "None"
	if len(settings.LoggingChannelID) > 0 {
		loggingChannelStatus = "<#" + settings.LoggingChannelID + ">"
	}

	err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Type:  discordgo.EmbedTypeRich,
					Title: "Logging Settings updated!",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Type",
							Value: settings.LogType.ToReadableString(),
						},
						{
							Name:  "Enabled",
							Value: strconv.FormatBool(settings.Enabled),
						},
						{
							Name:  "Channel",
							Value: loggingChannelStatus,
						},
						{
							Name:  "Format",
							Value: settings.Format,
						},
					},
					Color:     util.EmbedColorOK,
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
