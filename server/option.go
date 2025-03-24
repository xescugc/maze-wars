package server

var (
	Version     string = "dev"
	Environment string = "dev"
)

type Options struct {
	Port             string
	Verbose          bool
	Version          string
	DiscordBotToken  string
	DiscordChannelID string
}
