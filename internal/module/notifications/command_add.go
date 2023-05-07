package notifications

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	defaultNotificationFormats = map[GuildNotificationType]string{
		TwitchLive: "@everyone {twitch_name} just went live! {twitch_url}",
	}
)

func (m *Module) handleNotificationsAddCommand(interaction *discordgo.Interaction, optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	handleError := func(details string) {
		err := m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: details,
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
		}
	}

	if _, ok := optionMap[CommandOptionAddTwitchCmd]; ok {
		// Twitch
		channelID, ok := optionMap[CommandOptionChannel]
		if !ok || channelID.Type != discordgo.ApplicationCommandOptionChannel {
			handleError("Notification channel missing or invalid")
			return
		}

		twitchChannelName, ok := optionMap[CommandOptionTwitchChannel]
		if !ok || twitchChannelName.Type != discordgo.ApplicationCommandOptionString {
			handleError("Notification Twitch channel missing or invalid")
			return
		}

		var resultCount int64
		dbResult := m.db.Model(&GuildNotification{}).Where(&GuildNotification{GuildID: interaction.GuildID}).Count(&resultCount)
		if dbResult.Error != nil && !errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
			m.logger.Error("Error fetching notification settings from db", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(dbResult.Error))
			handleError("An internal error occurred. [" + interaction.ID + "]")
			return
		}
		if resultCount >= 20 {
			handleError("This Guild has reached the limit of 20 configured notifications.")
			return
		}

		users, err := m.twitch.GetUsers([]string{twitchChannelName.StringValue()})
		if err != nil {
			m.logger.Error("Error fetching twitch channel data", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
			handleError("There was an error fetching channel data from Twitch. [" + interaction.ID + "]")
			return
		}

		if len(users.Data) == 0 {
			handleError("The given channel was not found on Twitch.")
			return
		}

		twitchID := users.Data[0].ID

		notification := GuildNotification{
			GuildID:           interaction.GuildID,
			ChannelID:         channelID.ChannelValue(nil).ID,
			NotificationType:  TwitchLive,
			ProviderChannelID: twitchID,
			Format:            defaultNotificationFormats[TwitchLive],
		}

		err = m.db.Create(&notification).Error

		if err != nil {
			m.logger.Error("Error fetching twitch channel data", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
			handleError("An internal error occurred. [" + interaction.ID + "]")
			return
		}

		err = m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Notification added.",
			},
		})
		if err != nil {
			m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
		}
		return
	}
}
