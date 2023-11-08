package store

import (
	"bytes"
	"image"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/utils"
)

var (
	mapImages = map[int]image.Image{}
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
	mapImages[2] = ebiten.NewImageFromImage(m2)
	mapImages[3] = ebiten.NewImageFromImage(m3)
	mapImages[4] = ebiten.NewImageFromImage(m4)
	mapImages[5] = ebiten.NewImageFromImage(m5)
	mapImages[6] = ebiten.NewImageFromImage(m6)
}

// Map is a struct that holds all the information of the current map
type Map struct {
	*flux.ReduceStore

	store *Store
}

type MapState struct {
	Players int
	Image   image.Image
}

// NewMap initializes the map
func NewMap(d *flux.Dispatcher, s *Store) *Map {
	m := &Map{
		store: s,
	}
	m.ReduceStore = flux.NewReduceStore(d, m.Reduce, MapState{
		Players: 2,
		Image:   mapImages[2],
	})

	return m
}

// GetX returns the max X value of the map
func (m *Map) GetX() int { return m.GetState().(MapState).Image.Bounds().Dx() }

// GetY returns the max Y value of the map
func (m *Map) GetY() int { return m.GetState().(MapState).Image.Bounds().Dy() }

// GetNextLineID based on the map and max number of players
// it returns the next one and when it reaches the end
// then starts again
func (m *Map) GetNextLineID(clid int) int {
	clid += 1
	// For now as we only have 2 players
	// 0 is for player 1 and 1 for player 2
	// so anything higher that 1 has to go back
	// to 0
	// This should change depending on the
	// number of players on the game
	if clid > (m.GetState().(MapState).Players - 1) {
		clid = 0
	}
	return clid
}

// GetRandomSpawnCoordinatesForLineID returns from a lineID lid a random
// spawn coordinate to summon the units, it returns the X and Y coordinates
func (m *Map) GetRandomSpawnCoordinatesForLineID(lid int) (float64, float64) {
	// Starts at x:16,y:16, add it goes x*16 and y*7
	// then the next one is at x*10 and the same
	// The area is of 112

	p := rand.Intn(112)
	yy := (p%7)*16 + 16
	xx := ((p%16)*16 + 16) + (lid * 16 * (16 + 1 + 10 + 1))

	return float64(xx), float64(yy)
}

func (m *Map) GetHomeCoordinates(lid int) (float64, float64) {
	var y float64
	if lid > 3 {
		// 86 is the total length of the map including borders
		// and 10 is the vertical separation between the maps
		y = 86 + 10
		lid = -3
	}

	return float64(lid * 16 * (16 + 1 + 10 + 1)), y
}

func (m *Map) EndZone(lid int) utils.Object {
	return utils.Object{
		X: float64(16 + (lid * 16 * (16 + 1 + 10 + 1))),
		Y: 82 * 16,
		W: 16 * 16,
		H: 3 * 16,
	}
}

func (m *Map) BuildingZone(lid int) utils.Object {
	return utils.Object{
		X: float64(16 + (lid * 16 * (16 + 1 + 10 + 1))),
		Y: (7 * 16) + 16, // This +16 is for the border
		W: 16 * 16,
		H: 74 * 16,
	}
}

func (m *Map) UnitZone(lid int) utils.Object {
	return utils.Object{
		X: float64(16 + (lid * 16 * (16 + 1 + 10 + 1))),
		Y: 16, // This +16 is for the border
		W: 16 * 16,
		H: 81 * 16,
	}
}

// IsAtTheEnd checks if the Object obj on the lineID lid has reached the end of the
// line on it's position
func (m *Map) IsAtTheEnd(obj utils.Object, lid int) bool {
	return obj.IsColliding(m.EndZone(lid))
}

func (m *Map) IsInValidBuildingZone(obj utils.Object, lid int) bool {
	return m.BuildingZone(lid).IsInside(obj)
}

func (m *Map) IsInValidUnitZone(obj utils.Object, lid int) bool {
	return m.UnitZone(lid).IsInside(obj)
}

func (m *Map) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	mstate, ok := state.(MapState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.StartGame:
		players := m.store.Players.GetPlayers()
		if len(players) > 1 {
			allReady := true
			for _, p := range players {
				if !p.Ready {
					allReady = false
					break
				}
			}
			// TODO: This action could be done from the NavigateTo from the
			// lobby but we need the server to do the same logic and it does
			// not make sense to send NavigateTo events to the server.
			//
			// If all players are ready then the map must be set
			if allReady {
				mstate.Players = len(players)
				mstate.Image, ok = mapImages[mstate.Players]
				if !ok {
					log.Fatalf("The map for the number of players %d is not available", mstate.Players)
				}
			}
		}
	}

	return mstate
}
