package logging

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func (m *Module) registerMemberRoleListeners() {
	m.discord.AddHandler(m.handleMemberUpdate)
}

func (m *Module) handleMemberUpdate(_ *discordgo.Session, memberUpdate *discordgo.GuildMemberUpdate) {
	oldMember := memberUpdate.BeforeUpdate

	if added, removed, hasChanges := findRoleDifferences(oldMember.Roles, memberUpdate.Roles); hasChanges {
		guild, err := m.discord.State.Guild(memberUpdate.GuildID)
		var guildRoles []*discordgo.Role
		if err != nil {
			guildRoles, err = m.discord.GuildRoles(memberUpdate.GuildID)
			if err != nil {
				m.sendErrorLogToDiscord(memberUpdate.GuildID, MemberRoleChange, "Roles of member "+memberUpdate.User.String()+" changed, but guild roles could not be fetched.")
				return
			}
		} else {
			guildRoles = guild.Roles
		}
		oldRoleNames := mapRoleIDsToRoleNames(guildRoles, oldMember.Roles)
		newRoleNames := mapRoleIDsToRoleNames(guildRoles, memberUpdate.Roles)

		addedRoleNames := mapRoleIDsToRoleNames(guildRoles, added)
		removedRoleNames := mapRoleIDsToRoleNames(guildRoles, removed)

		var roleChanges []string
		for _, ar := range addedRoleNames {
			roleChanges = append(roleChanges, "+"+ar)
		}
		for _, rr := range removedRoleNames {
			roleChanges = append(roleChanges, "-"+rr)
		}

		m.sendLogToDiscord(memberUpdate.GuildID, MemberRoleChange, map[string]string{
			"member_id":        memberUpdate.User.ID,
			"member_full_name": memberUpdate.User.String(),
			"old_roles":        strings.Join(oldRoleNames, ","),
			"new_roles":        strings.Join(newRoleNames, ","),
			"role_changes":     strings.Join(roleChanges, ","),
		})
	}
}

func findRoleDifferences(oldRoles []string, newRoles []string) ([]string, []string, bool) {
	var removed []string

	for _, oldRole := range oldRoles {
		found := false
		for i, newRole := range newRoles {
			if oldRole == newRole {
				newRoles = append(newRoles[:i], newRoles[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, oldRole)
		}
	}

	return newRoles, removed, len(newRoles)+len(removed) > 0
}

func mapRoleIDsToRoleNames(roles []*discordgo.Role, ids []string) []string {
	var names []string
	for i, id := range ids {
		for _, role := range roles {
			if role.ID == id {
				names = append(names, role.Name)
				break
			}
		}
		if len(names) <= i {
			names = append(names, "Unknown Role")
		}
	}
	return names
}
