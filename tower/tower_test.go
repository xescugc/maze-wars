package tower_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit/environment"
)

func TestTowerKeyFunctions(t *testing.T) {
	tw := tower.Towers[tower.Range1.String()]
	t.Run("FacesetKey", func(t *testing.T) {
		assert.Equal(t, "t-f-range1", tw.FacesetKey())
	})
	t.Run("IdleKey", func(t *testing.T) {
		assert.Equal(t, "t-i-range1", tw.IdleKey())
	})
	t.Run("ProfileKey", func(t *testing.T) {
		assert.Equal(t, "t-p-range1", tw.ProfileKey())
	})
}

func TestTowerCanTarget(t *testing.T) {
	tw := tower.Towers[tower.Melee1.String()]
	assert.True(t, tw.CanTarget(environment.Terrestrial))
	assert.False(t, tw.CanTarget(environment.Aerial))
}

func TestTowerShootsArrows(t *testing.T) {
	tw := tower.Towers[tower.RangeAoE1.String()]
	assert.False(t, tw.ShootsArrows())

	tw = tower.Towers[tower.Melee1.String()]
	assert.True(t, tw.ShootsArrows())
}

func TestTowerName(t *testing.T) {
	pt := tower.Towers[tower.Range1.String()]
	tw := *pt
	assert.Equal(t, "Range - T1", tw.Name())

	tw.Type = tower.Type(100)
	assert.Equal(t, "Type(100)", tw.Name())
}

func TestTowerDescription(t *testing.T) {
	pt := tower.Towers[tower.Range1.String()]
	tw := *pt
	assert.Equal(t, "Basic range tower", tw.Description())

	tw.Type = tower.Type(100)
	assert.Equal(t, "Type(100)", tw.Description())
}
