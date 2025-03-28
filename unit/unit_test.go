package unit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
)

func TestUnitKeyFunctions(t *testing.T) {
	u := unit.Unit{Type: unit.Ninja}
	t.Run("FacesetKey", func(t *testing.T) {
		assert.Equal(t, "u-f-ninja", u.FacesetKey())
	})
	t.Run("WalkKey", func(t *testing.T) {
		assert.Equal(t, "u-w-ninja", u.WalkKey())
	})
	t.Run("AttackKey", func(t *testing.T) {
		assert.Equal(t, "u-a-ninja", u.AttackKey())
	})
	t.Run("IdleKey", func(t *testing.T) {
		assert.Equal(t, "u-i-ninja", u.IdleKey())
	})
	t.Run("ProfileKey", func(t *testing.T) {
		assert.Equal(t, "u-p-ninja", u.ProfileKey())
	})
}

func TestUnitHasAbility(t *testing.T) {
	u := unit.Unit{
		Type:      unit.Ninja,
		Abilities: []ability.Ability{ability.Efficiency},
	}
	assert.True(t, u.HasAbility(ability.Efficiency))
	assert.False(t, u.HasAbility(ability.Burrow))
}

func TestUnitName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := unit.Unit{Type: unit.Ninja}
		assert.Equal(t, "Ninja", u.Name())
	})
	t.Run("Error", func(t *testing.T) {
		u := unit.Unit{Type: unit.Type(100)}
		assert.Equal(t, "Type(100)", u.Name())
	})
}
