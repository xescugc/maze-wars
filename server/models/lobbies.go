package models

type LobbiesResponse struct {
	Lobbies []LobbyResponse `json:"lobbies"`
}

type LobbyResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MaxPlayers int    `json:"max_players"`

	// Players holds the usernames
	// including the owner one
	// The value is if it's a bot or not
	Players map[string]bool `json:"players"`

	// The username of the owner
	Owner string `json:"owner"`
}
