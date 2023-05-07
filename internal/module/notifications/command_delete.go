package notifications

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

func (m *Module) handleNotificationsDeleteCommand(interaction *discordgo.Interaction, optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	handleError := func(details string) {
		m.logger.Error("Error parsing notification update", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.String("details", details))
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

	notificationIDOption, ok := optionMap[CommandOptionDeleteId]
	if !ok || notificationIDOption.Type != discordgo.ApplicationCommandOptionString {
		handleError("Notification Twitch channel missing or invalid")
		return
	}
	notificationIDStr := notificationIDOption.StringValue()
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		handleError("Invalid notification ID.")
		return
	}

	dbResult := m.db.Delete(&GuildNotification{GuildID: interaction.GuildID, ID: notificationID})

	if dbResult.Error != nil && !errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		m.logger.Error("Error fetching notification settings from db", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(dbResult.Error))
		handleError("An internal error occurred. [" + interaction.ID + "]")
		return
	}

	if dbResult.RowsAffected == 0 {
		handleError("The given notification ID was not found on your guild.")
		return
	}

	err = m.discord.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Notification with ID " + notificationIDStr + " was deleted.",
		},
	})
	if err != nil {
		m.logger.Error("Error responding to interaction", zap.String("guild", interaction.GuildID), zap.String("interaction", interaction.ID), zap.Error(err))
	}

}
