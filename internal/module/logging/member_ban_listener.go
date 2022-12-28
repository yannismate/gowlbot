package logging

import (
	"github.com/bwmarrin/discordgo"
)

func (m *Module) registerMemberBanListeners() {
	m.discord.AddHandler(m.handleGuildBanAdd)
	m.discord.AddHandler(m.handleGuildBanRemove)
}

func (m *Module) handleGuildBanAdd(_ *discordgo.Session, banAdd *discordgo.GuildBanAdd) {
	m.sendLogToDiscord(banAdd.GuildID, GuildBanAdd, map[string]string{
		"member_id":        banAdd.User.ID,
		"member_full_name": banAdd.User.String(),
	})
}

func (m *Module) handleGuildBanRemove(_ *discordgo.Session, banRemove *discordgo.GuildBanRemove) {
	m.sendLogToDiscord(banRemove.GuildID, GuildBanRemove, map[string]string{
		"member_id":        banRemove.User.ID,
		"member_full_name": banRemove.User.String(),
	})
}
