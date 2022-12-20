package util

import "github.com/bwmarrin/discordgo"

func ExtractOptionsMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)

	for _, option := range options {
		if _, ok := optionMap[option.Name]; !ok || option.Value != nil {
			optionMap[option.Name] = option
		}

		if len(option.Options) > 0 {
			subMap := ExtractOptionsMap(option.Options)
			for k, v := range subMap {
				optionMap[k] = v
			}
		}
	}

	return optionMap
}
