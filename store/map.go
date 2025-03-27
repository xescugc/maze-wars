package store

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/assets"
)

var (
	MapImages = map[int]image.Image{}

	MapImageKeyFmt = "m-%d"
)

func init() {
	m2, _, err := image.Decode(bytes.NewReader(assets.Map_2_png))
	if err != nil {
		log.Fatal(err)
	}
	m3, _, err := image.Decode(bytes.NewReader(assets.Map_3_png))
	if err != nil {
		log.Fatal(err)
	}
	m4, _, err := image.Decode(bytes.NewReader(assets.Map_4_png))
	if err != nil {
		log.Fatal(err)
	}
	m5, _, err := image.Decode(bytes.NewReader(assets.Map_5_png))
	if err != nil {
		log.Fatal(err)
	}
	m6, _, err := image.Decode(bytes.NewReader(assets.Map_6_png))
	if err != nil {
		log.Fatal(err)
	}
	MapImages[2] = m2
	MapImages[3] = m3
	MapImages[4] = m4
	MapImages[5] = m5
	MapImages[6] = m6
}

// Map is a struct that holds all the information of the current map
type Map struct {
	*flux.ReduceStore[MapState, *action.Action]

	store *Store
}

type MapState struct {
	Players int
	Image   image.Image
}

// NewMap initializes the map
func NewMap(d *flux.Dispatcher[*action.Action], s *Store) *Map {
	m := &Map{
		store: s,
	}
	m.ReduceStore = flux.NewReduceStore(d, m.Reduce, MapState{
		Players: 2,
		Image:   MapImages[2],
	})

	return m
}

// GetX returns the max X value of the map
func (m *Map) GetX() int { return m.GetState().Image.Bounds().Dx() }

// GetY returns the max Y value of the map
func (m *Map) GetY() int { return m.GetState().Image.Bounds().Dy() }

// GetY returns the max Y value of the map
func (m *Map) GetImageKey() string {
	pc := m.GetState().Players
	return fmt.Sprintf(MapImageKeyFmt, pc)
}

// GetNextLineID based on the map and max number of players
// it returns the next one and when it reaches the end
// then starts again
func (m *Map) GetNextLineID(clid int) int {
	clid += 1
	if clid > (m.GetState().Players - 1) {
		clid = 0
	}
	return clid
}

func (m *Map) GetHomeCoordinates(lid int) (int, int) {
	return lid*16*(16+1+10+1) + (43 * 16), (43 * 16)
}

func (m *Map) Reduce(state MapState, act *action.Action) MapState {
	switch act.Type {
	case action.StartGame:
		m.GetDispatcher().WaitFor(m.store.Game.GetDispatcherToken())

		var ok bool
		state.Players = len(m.store.Game.ListPlayers())
		state.Image, ok = MapImages[state.Players]
		if !ok {
			log.Fatalf("The map for the number of players %d is not available", state.Players)
		}
	}

	return state
}
