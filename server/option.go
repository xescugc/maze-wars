package server

var (
	Version          string = "dev"
	Environment      string = "dev"
	DiscordBotToken  string = "non"
	DiscordChannelID string = "non"
)

type Options struct {
	Port             string
	Verbose          bool
	Version          string
	DiscordBotToken  string
	DiscordChannelID string
}
