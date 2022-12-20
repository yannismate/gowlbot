package logging

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strconv"
)

func (m *Module) registerMemberJoinLeaveListeners() {
	m.discord.AddHandler(m.handleMemberJoin)
	m.discord.AddHandler(m.handleMemberLeave)
}

func (m *Module) handleMemberJoin(_ *discordgo.Session, add *discordgo.GuildMemberAdd) {
	guild, err := m.discord.State.Guild(add.GuildID)
	if err != nil {
		m.logger.Error("Error getting guild state to log member join", zap.String("guild", add.GuildID), zap.Error(err))
		m.sendLogToDiscord(add.GuildID, MemberJoin, map[string]string{
			"member_id":          add.User.ID,
			"member_full_name":   add.User.String(),
			"guild_member_count": "Unknown",
		})
		return
	}
	memberCount := strconv.Itoa(guild.MemberCount)

	m.sendLogToDiscord(add.GuildID, MemberJoin, map[string]string{
		"member_id":          add.User.ID,
		"member_full_name":   add.User.String(),
		"guild_member_count": memberCount,
	})
}

func (m *Module) handleMemberLeave(_ *discordgo.Session, remove *discordgo.GuildMemberRemove) {
	guild, err := m.discord.State.Guild(remove.GuildID)
	if err != nil {
		m.logger.Error("Error getting guild state to log member leave", zap.String("guild", remove.GuildID), zap.Error(err))
		m.sendLogToDiscord(remove.GuildID, MemberLeave, map[string]string{
			"member_id":          remove.User.ID,
			"member_full_name":   remove.User.String(),
			"guild_member_count": "Unknown",
		})
		return
	}
	memberCount := strconv.Itoa(guild.MemberCount)

	m.sendLogToDiscord(remove.GuildID, MemberLeave, map[string]string{
		"member_id":          remove.User.ID,
		"member_full_name":   remove.User.String(),
		"guild_member_count": memberCount,
	})
}
