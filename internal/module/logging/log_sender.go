package logging

import (
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func (m *Module) sendLogToDiscord(guildID string, logType LogType, data map[string]string) {

	data["time"] = strconv.FormatInt(time.Now().UnixMilli()/1000, 10)

	logSettings := GuildLoggingSetting{}

	result := m.db.Where(&GuildLoggingSetting{GuildID: guildID, LogType: logType}).First(&logSettings)

	if result.Error != nil || result.RowsAffected == 0 || !logSettings.Enabled || len(logSettings.LoggingChannelID) == 0 {
		return
	}

	m.logger.Debug("Logging Event",
		zap.Any("guild", guildID),
		zap.Any("channel", logSettings.LoggingChannelID),
		zap.Any("logType", logType),
		zap.Any("data", data),
	)

	var secondMessageContent string

	var replaceList []string
	for key, value := range data {
		if key == "previous_content" && strings.Contains(logSettings.Format, "previous_content") {
			if utf8.RuneCountInString(value) > 1000 {
				fullValue := value
				value = escapeDiscordString(substringUTF8(fullValue, 0, 1000))
				secondMessageContent = escapeDiscordString(substringUTF8(fullValue, 1000, 2000))
			} else {
				value = escapeDiscordString(value)
			}
		}
		replaceList = append(replaceList, "{"+key+"}", value)
	}
	replacer := strings.NewReplacer(replaceList...)

	resultString := replacer.Replace(logSettings.Format)

	_, err := m.discord.ChannelMessageSend(logSettings.LoggingChannelID, resultString)
	if err != nil {
		m.logger.Error("Error sending log message to Discord", zap.Any("guild", guildID), zap.Any("channel", logSettings.LoggingChannelID), zap.Error(err))
		return
	}

	if len(secondMessageContent) > 0 {
		_, err = m.discord.ChannelMessageSend(logSettings.LoggingChannelID, secondMessageContent)
		if err != nil {
			m.logger.Error("Error sending second log message to Discord", zap.Any("guild", guildID), zap.Any("channel", logSettings.LoggingChannelID), zap.Error(err))
		}
	}
}

func (m *Module) sendErrorLogToDiscord(guildID string, logType LogType, message string) {

	logSettings := GuildLoggingSetting{}

	result := m.db.Where(&GuildLoggingSetting{GuildID: guildID, LogType: logType}).First(&logSettings)

	if result.Error != nil || result.RowsAffected == 0 {
		return
	}

	m.logger.Debug("Logging Error Event",
		zap.Any("guild", guildID),
		zap.Any("logType", logType),
		zap.Any("message", message),
	)

	timestamp := strconv.FormatInt(time.Now().UnixMilli()/1000, 10)
	_, err := m.discord.ChannelMessageSend(logSettings.LoggingChannelID, "<t:"+timestamp+"> Internal Error: "+message)
	if err != nil {
		m.logger.Error("Error sending error log message to Discord", zap.Any("guild", guildID), zap.Any("channel", logSettings.LoggingChannelID), zap.Error(err))
	}
}

func substringUTF8(s string, start int, end int) string {
	startStrIdx := 0
	i := 0
	for j := range s {
		if i == start {
			startStrIdx = j
		}
		if i == end {
			return s[startStrIdx:j]
		}
		i++
	}
	return s[startStrIdx:]
}

func escapeDiscordString(content string) string {
	return "```" + strings.ReplaceAll(content, "`", "`\u200B") + "```"
}
