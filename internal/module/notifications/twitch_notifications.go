package notifications

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/yannismate/gowlbot/internal/util"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"strings"
	"time"
)

func (m *Module) startTwitchUpdateTimer() {
	go func() {
		for range time.Tick(time.Minute) {
			var notifications []GuildNotification
			dbRes := m.db.Where(&GuildNotification{NotificationType: TwitchLive}).Find(&notifications)

			if dbRes.Error != nil {
				m.logger.Error("Error while fetching twitch server notifications from DB", zap.Error(dbRes.Error))
				continue
			}

			notificationsByTwitchChannelID := make(map[string][]GuildNotification)
			for _, nt := range notifications {
				if oldVal, ok := notificationsByTwitchChannelID[nt.ProviderChannelID]; ok {
					notificationsByTwitchChannelID[nt.ProviderChannelID] = append(oldVal, nt)
				} else {
					notificationsByTwitchChannelID[nt.ProviderChannelID] = []GuildNotification{nt}
				}
			}

			chunked := util.ChunkMap(notificationsByTwitchChannelID, 100)

			for _, chunk := range chunked {
				twitchRes, err := m.twitch.GetStreams(maps.Keys(chunk))
				if err != nil {
					m.logger.Error("Error while fetching streams from Twitch", zap.Error(err))
					continue
				}
				for _, streamData := range twitchRes.Data {
					if streamData.Type != "live" {
						continue
					}
					if m.isNewlyLive(streamData.UserID, streamData.StartedAt) {
						if notificationsForTwitchChannel, ok := chunk[streamData.UserID]; ok {
							for _, nt := range notificationsForTwitchChannel {
								m.sendTwitchNotification(nt, streamData.UserName, streamData.GameName, streamData.Title, streamData.UserLogin, streamData.ThumbnailURL, streamData.StartedAt)
							}
						}
					}
				}
			}
		}
	}()
}

func (m *Module) isNewlyLive(twitchUserID string, startedAt time.Time) bool {
	if startedAt.Before(time.Now().Add(-5 * time.Minute)) {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	numExists, err := m.cache.Exists(ctx, "twitch-live:"+twitchUserID).Result()
	if err != nil {
		m.logger.Error("Error getting stream status from cache", zap.Error(err), zap.String("twitchUser", twitchUserID))
		return false
	}
	foundInCache := numExists > 0

	// channel has to be offline for 5 minutes for notification to be triggered again
	err = m.cache.Set(ctx, "twitch-live:"+twitchUserID, "", time.Minute*5).Err()
	if err != nil {
		m.logger.Error("Error updating stream status in cache", zap.Error(err), zap.String("twitchUser", twitchUserID))
	}

	return !foundInCache
}

func (m *Module) sendTwitchNotification(notification GuildNotification, userName string, gameName string, title string, userLogin string, thumbnailURL string, startedAt time.Time) {
	twitchURL := "https://twitch.tv/" + userLogin

	thumbnailURLReplacer := strings.NewReplacer("{width}", "1280", "{height}", "720")

	replaceList := []string{
		"{twitch_name}", userName,
		"{twitch_url}", twitchURL,
		"{twitch_game_name}", gameName,
		"{twitch_title}", title,
		"{twitch_thumbnail_url}", thumbnailURLReplacer.Replace(thumbnailURL),
	}
	replacer := strings.NewReplacer(replaceList...)

	msg := discordgo.MessageSend{
		Content: replacer.Replace(notification.Format),
		Embed: &discordgo.MessageEmbed{
			Title: userName + " - Twitch",
			URL:   twitchURL,
			Type:  discordgo.EmbedTypeRich,
			Image: &discordgo.MessageEmbedImage{
				URL: thumbnailURLReplacer.Replace(thumbnailURL),
			},
			Timestamp: startedAt.Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "gowlbot " + util.GetVersionString(),
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Title",
					Value: title,
				},
				{
					Name:  "Game",
					Value: gameName,
				},
			},
		},
	}
	_, err := m.discord.ChannelMessageSendComplex(notification.ChannelID, &msg)
	if err != nil {
		m.logger.Warn("Error sending twitch notification message", zap.String("guild", notification.GuildID), zap.String("channel", notification.ChannelID), zap.Error(err))
	}
}
