package main

import (
	"github.com/yannismate/gowlbot/internal/config"
	"github.com/yannismate/gowlbot/internal/db"
	"github.com/yannismate/gowlbot/internal/discord"
	"github.com/yannismate/gowlbot/internal/module"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {

	var providers []interface{}

	providers = append(providers, zap.NewProduction)
	providers = append(providers, config.ProvideConfig)
	providers = append(providers, db.ProvideDB)
	providers = append(providers, discord.ProvideDiscordClient)
	providers = append(providers, module.GetRegisteredModules()...)

	fx.New(
		fx.Provide(
			providers...,
		),
		fx.Invoke(module.StartModules),
	).Run()

}
