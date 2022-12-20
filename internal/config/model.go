package config

type OwlBotConfig struct {
	Discord DiscordConfig `yaml:"discord"`
	Cache   CacheConfig   `yaml:"cache"`
}

type DiscordConfig struct {
	BotToken      string `yaml:"bot-token"`
	ApplicationID string `yaml:"application-id"`
}

type CacheConfig struct {
	URL               string `yaml:"url"`
	MessageTTLMinutes int    `yaml:"message-ttl-minutes"`
}
