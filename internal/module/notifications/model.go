package notifications

type GuildNotificationType string

const (
	TwitchLive GuildNotificationType = "twitch_live"
)

func (gnt *GuildNotificationType) ToString() string {
	switch *gnt {
	case TwitchLive:
		return "Twitch"
	}
	return "Unknown"
}

type GuildNotification struct {
	ID                int64 `gorm:"primaryKey;autoIncrement;not_null"`
	GuildID           string
	ChannelID         string
	NotificationType  GuildNotificationType
	ProviderChannelID string
	Format            string
}
