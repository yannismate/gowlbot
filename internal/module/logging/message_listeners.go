package logging

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"sort"
	"time"
)

func (m *Module) registerMessageListeners() {
	m.discord.AddHandler(m.handleMessageCreation)
	m.discord.AddHandler(m.handleMessageDeletion)
	m.discord.AddHandler(m.handleMessageBulkDeletion)
	m.discord.AddHandler(m.handleMessageEdit)
}

func (m *Module) handleMessageCreation(_ *discordgo.Session, msg *discordgo.MessageCreate) {
	m.storeMessageInCache(msg.Message)
}

func (m *Module) handleMessageDeletion(_ *discordgo.Session, msg *discordgo.MessageDelete) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cachedMsg := CachedMessage{}

	err := m.cache.Get(ctx, "discord-msg:"+msg.ID).Scan(&cachedMsg)
	if err != nil {
		errorMsg := "<#" + msg.ChannelID + "> Message with ID *" + msg.ID + "* was deleted but the content could not be found in the bots cache."
		m.sendErrorLogToDiscord(msg.GuildID, MessageDelete, errorMsg)
		return
	}

	m.sendLogToDiscord(msg.GuildID, MessageDelete, map[string]string{
		"channel_id":       msg.ChannelID,
		"author_id":        cachedMsg.AuthorID,
		"author_full_name": cachedMsg.AuthorFullName,
		"previous_content": cachedMsg.Content,
	})
}

func (m *Module) handleMessageBulkDeletion(_ *discordgo.Session, msgBulk *discordgo.MessageDeleteBulk) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// make sure messages are being logged in order
	sortedIds := msgBulk.Messages
	sort.Strings(sortedIds)

	for _, msgID := range sortedIds {

		cachedMsg := CachedMessage{}

		err := m.cache.Get(ctx, "discord-msg:"+msgID).Scan(&cachedMsg)
		if err != nil {
			errorMsg := "<#" + msgBulk.ChannelID + "> Message with ID *" + msgID + "* was deleted but the content could not be found in the bots cache."
			m.sendErrorLogToDiscord(msgBulk.GuildID, MessageDelete, errorMsg)
			continue
		}

		m.sendLogToDiscord(msgBulk.GuildID, MessageDelete, map[string]string{
			"channel_id":       msgBulk.ChannelID,
			"author_id":        cachedMsg.AuthorID,
			"author_full_name": cachedMsg.AuthorFullName,
			"previous_content": cachedMsg.Content,
		})
	}
}

func (m *Module) handleMessageEdit(_ *discordgo.Session, msg *discordgo.MessageUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cachedMsg := CachedMessage{}

	err := m.cache.Get(ctx, "discord-msg:"+msg.ID).Scan(&cachedMsg)
	if err != nil {
		errorMsg := "<#" + msg.ChannelID + "> Message with ID *" + msg.ID + "* sent by *" + msg.Author.String() + "* was edited but the previous content could not be found in the bots cache."
		m.sendErrorLogToDiscord(msg.GuildID, MessageEdit, errorMsg)
		return
	}

	if msg.Author != nil {
		m.sendLogToDiscord(msg.GuildID, MessageEdit, map[string]string{
			"channel_id":       msg.ChannelID,
			"author_id":        msg.Author.ID,
			"author_full_name": msg.Author.String(),
			"previous_content": cachedMsg.Content,
		})
	} else if len(msg.WebhookID) > 0 {
		m.sendLogToDiscord(msg.GuildID, MessageEdit, map[string]string{
			"channel_id":       msg.ChannelID,
			"author_id":        msg.WebhookID,
			"author_full_name": "Webhook",
			"previous_content": cachedMsg.Content,
		})
	} else {
		if len(msg.Embeds) > 0 {
			return
		}
		m.logger.Error("Message did not have user or webhook id!", zap.Any("message", msg))
		return
	}

	m.storeMessageInCache(msg.Message)
}

func (m *Module) storeMessageInCache(msg *discordgo.Message) {
	var cachedMsg CachedMessage

	if msg.Author != nil {
		cachedMsg = CachedMessage{
			ChannelID:      msg.ChannelID,
			AuthorID:       msg.Author.ID,
			AuthorFullName: msg.Author.String(),
			Content:        msg.Content,
		}
	} else if len(msg.WebhookID) > 0 {
		cachedMsg = CachedMessage{
			ChannelID:      msg.ChannelID,
			AuthorID:       msg.WebhookID,
			AuthorFullName: "Webhook",
			Content:        msg.Content,
		}
	} else {
		if len(msg.Embeds) > 0 {
			return
		}
		m.logger.Error("Message did not have user or webhook id!", zap.Any("message", msg))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ttl := time.Minute * time.Duration(m.config.Cache.MessageTTLMinutes)

	err := m.cache.Set(ctx, "discord-msg:"+msg.ID, &cachedMsg, ttl).Err()
	if err != nil {
		m.logger.Warn("Error storing message in cache", zap.Error(err))
	}
}
