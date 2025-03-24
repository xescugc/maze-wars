package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	currentNumberOfPlayers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "maze_wars",
			Name:      "current_players",
		},
	)

	currentNumberOfGames = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "maze_wars",
			Name:      "current_games",
		},
	)

	numberOfGames = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "maze_wars",
			Name:      "number_of_games_count",
		},
		[]string{
			"type",
		},
	)

	numberOfActions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "maze_wars",
			Name:      "number_of_actions_count",
		},
		[]string{
			"type",
		},
	)
)
