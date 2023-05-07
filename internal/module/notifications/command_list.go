package notifications

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func (m *Module) handleNotificationsListCommand(interaction *discordgo.Interaction) {

	var notifications []GuildNotification

	result := m.db.Where(&GuildNotification{GuildID: interaction.GuildID}).Find(&notifications)

	if result.Error != nil {
		m.logger.Error("Error fetching guild notification settings", zap.Any("guild", interaction.GuildID), zap.Any("interaction", interaction.ID), zap.Error(result.Error))
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "There was an error fetching the notification settings for this server.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.Any("guild", interaction.GuildID), zap.Any("interaction", interaction.ID), zap.Error(err))
		}
		return
	}

	notificationsByChannel := make(map[string][]string)
	for _, nt := range notifications {
		arr, ok := notificationsByChannel[nt.ChannelID]
		if !ok {
			arr = []string{}
		}
		arr = append(arr, strconv.FormatInt(nt.ID, 10)+": "+nt.NotificationType.ToString()+" - Channel ID "+nt.ProviderChannelID)
		notificationsByChannel[nt.ChannelID] = arr
	}

	var embedFields []*discordgo.MessageEmbedField
	for channel, nts := range notificationsByChannel {
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:  "<#" + channel + ">",
			Value: strings.Join(nts, "\n"),
		})
	}
	if len(embedFields) == 0 {
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name: "No notifications configured",
		})
	}

	err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Type:      discordgo.EmbedTypeRich,
					Title:     "Current Notification Status",
					Fields:    embedFields,
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
