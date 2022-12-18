package main

import (
	"github.com/yannismate/gowlbot/internal/config"
	"github.com/yannismate/gowlbot/internal/db"
	"github.com/yannismate/gowlbot/internal/discord"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	
	
	fx.New(fx.Provide(
		zap.NewProduction,
		config.ProvideConfig,
		db.ProvideDB,
		discord.ProvideDiscordClient,
	)).Run()

}