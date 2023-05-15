package config

type OwlBotConfig struct {
	Discord DiscordConfig `yaml:"discord"`
	Cache   CacheConfig   `yaml:"cache"`
	Twitch  TwitchConfig  `yaml:"twitch"`
}

type DiscordConfig struct {
	BotToken      string `yaml:"bot-token"`
	ApplicationID string `yaml:"application-id"`
}

type CacheConfig struct {
	URL               string `yaml:"url"`
	MessageTTLMinutes int    `yaml:"message-ttl-minutes"`
}

type TwitchConfig struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
}
