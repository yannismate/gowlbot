package logging

import (
	"encoding/json"
)

type CachedMessage struct {
	ChannelID      string
	AuthorID       string
	AuthorFullName string
	Content        string
}

func (cm *CachedMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(cm)
}

func (cm *CachedMessage) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, cm)
}

type LogType string

const (
	MessageEdit      LogType = "message_edit"
	MessageDelete    LogType = "message_delete"
	MemberJoin       LogType = "member_join"
	MemberLeave      LogType = "member_leave"
	MemberRoleChange LogType = "member_role_change"
	GuildBanAdd      LogType = "guild_ban_add"
	GuildBanRemove   LogType = "guild_ban_remove"
)

var (
	logTypeReadableStringsMap = map[LogType]string{
		MessageEdit:      "Message Edit",
		MessageDelete:    "Message Delete",
		MemberJoin:       "Member Join",
		MemberLeave:      "Member Leave",
		MemberRoleChange: "Member Role Change",
		GuildBanAdd:      "User Banned",
		GuildBanRemove:   "User Unbanned",
	}
	logTypeParseMap = map[string]LogType{
		"message_edit":       MessageEdit,
		"message_delete":     MessageDelete,
		"member_join":        MemberJoin,
		"member_leave":       MemberLeave,
		"member_role_change": MemberRoleChange,
		"guild_ban_add":      GuildBanAdd,
		"guild_ban_remove":   GuildBanRemove,
	}
)

func (lt LogType) ToReadableString() string {
	return logTypeReadableStringsMap[lt]
}

func ParseLogType(str string) (LogType, bool) {
	v, ok := logTypeParseMap[str]
	return v, ok
}

type GuildLoggingSetting struct {
	ID               uint    `gorm:"primaryKey"`
	GuildID          string  `gorm:"uniqueIndex:logging_server_type_idx"`
	LogType          LogType `gorm:"uniqueIndex:logging_server_type_idx"`
	Enabled          bool
	LoggingChannelID string
	Format           string
}
