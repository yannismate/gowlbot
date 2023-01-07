package notifications

import "github.com/bwmarrin/discordgo"

func (m *Module) MigrateSlashCommands(guildID string, oldCommands []*discordgo.ApplicationCommand) error {
	return nil
}
