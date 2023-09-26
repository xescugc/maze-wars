package store

import (
	"bytes"
	"image"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/utils"
)

// Map is a struct that holds all the information of the current map
type Map struct {
	Image image.Image
}

// NewMap initializes the map
func NewMap() (*Map, error) {
	mi, _, err := image.Decode(bytes.NewReader(assets.M_1v1_png))
	if err != nil {
		return nil, err
	}

	return &Map{
		Image: ebiten.NewImageFromImage(mi),
	}, nil
}

// GetX returns the max X value of the map
func (m *Map) GetX() int { return m.Image.Bounds().Dx() }

// GetY returns the max Y value of the map
func (m *Map) GetY() int { return m.Image.Bounds().Dy() }

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
	if clid > 1 {
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

	//p := rand.Intn(15 * 16 * 6 * 16)
	//yy := (p % (6 * 16)) + 16
	//xx := (p % (15 * 16)) + 16

	return float64(xx), float64(yy)
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
