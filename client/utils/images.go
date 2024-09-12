package utils

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
)

var (
	Images *ImagesCache
)

const (
	CrossImageKey = "cross-image"
	ArrowImageKey = "arrow"

	BuffBurrowedKey      = "buff-burrowed"
	BuffBurrowedReadyKey = "buff-burrowed-ready"

	BuffResurrectingKey = "buff-resurrecting"

	LifeBarProgressKey   = "life-bar-progress"
	LifeBarUnderKey      = "life-bar-under"
	ShieldBarProgressKey = "shield-bar-progress"

	LifeBarBigProgressKey = "life-bar-big-progress"
	LifeBarBigUnderKey    = "life-bar-big-under"

	BGKey                = "bg"
	ButtonPressedKey     = "button-pressed"
	ButtonNormalKey      = "button-normal"
	ButtonHoverKey       = "button-hover"
	ButtonDisabledKey    = "button-disabled"
	BigButtonPressedKey  = "big-button-pressed"
	BigButtonNormalKey   = "big-button-normal"
	BigButtonHoverKey    = "big-button-hover"
	BigButtonDisabledKey = "big-button-disabled"
	MenuButtonPressedKey = "menu-button-pressed"
	MenuButtonHoverKey   = "menu-button-hover"

	LogoKey = "logo"

	Border4Key        = "border-4"
	SetupGameBGKey    = "setup-game-bg"
	SetupGameFrameKey = "setup-game-frame"

	BigButtonPressedLeftTabKey   = "big-button-pressed-left-tab"
	BigButtonNormalLeftTabKey    = "big-button-normal-left-tab"
	BigButtonHoverLeftTabKey     = "big-button-hover-left-tab"
	BigButtonDisabledLeftTabKey  = "big-button-disabled-left-tab"
	BigButtonPressedRightTabKey  = "big-button-pressed-right-tab"
	BigButtonNormalRightTabKey   = "big-button-normal-right-tab"
	BigButtonHoverRightTabKey    = "big-button-hover-right-tab"
	BigButtonDisabledRightTabKey = "big-button-disabled-right-tab"

	LobbiesTableKey = "lobbies-table"

	DarkInputBGKey = "dark-input-bg"
	GrayInputBGKey = "gray-input-bg"

	HSliderGrabberKey         = "h-slider-grabber"
	HSliderGrabberHoverKey    = "h-slider-grabber-hover"
	HSliderGrabberDisabledKey = "h-slider-grabber-disabled"

	CheckboxCheckedKey   = "checkbox-checked"
	CheckboxUncheckedKey = "checkbox-unchecked"

	BigCancelButtonNormalKey  = "big-cancel-button-normal"
	BigCancelButtonHoverKey   = "big-cancel-button-hover"
	BigCancelButtonPressedKey = "big-cancel-button-pressed"

	DisplayDefaultBGKey       = "display-default-bg"
	DisplayDefaultTowersBGKey = "display-default-towers-bg"

	ButtonBorderNormalKey   = "button-border-normal"
	ButtonBorderPressedKey  = "button-border-pressed"
	ButtonBorderHoverKey    = "button-border-hover"
	ButtonBorderDisabledKey = "button-border-disabled"

	ScoreboardRowBGKey        = "scoreboard-row-bg"
	ScoreboardRowCurrentBGKey = "scoreboard-row-current-bg"

	GoldIconKey        = "gold-icon"
	CapIconKey         = "cap-icon"
	IncomeIconKey      = "income-icon"
	IncomeTimerIconKey = "income-timer-icon"
	HeartIconKey       = "heart-icon"

	InfoLeftBGKey   = "info-left-bg"
	InfoMiddleBGKey = "info-middle-bg"
	InfoRightBGKey  = "info-right-bg"

	UnitUpdateButtonAnimationKey = "unit-update-button-animation"

	SellIconKey = "sell-icon"

	DisplayTargetImageBGKey   = "display-target-image-bg"
	DisplayTargetDetailsBGKey = "display-target-details-bg"

	DamageIconKey        = "damage-icon"
	AttackSpeedIconKey   = "attack-speed-icon"
	RangeIconKey         = "range-icon"
	MovementSpeedIconKey = "movement-speed-icon"
	PlayerIconKey        = "player-icon"

	ToolTipBGKey = "tooltip-bg"
)

// ImagesCache is a simple cache for all the images, so instead
// of running 'ebiten.NewImageFromImage' we just ran it once
// and reuse it all the time
type ImagesCache struct {
	images map[string]*ebiten.Image
}

// Get will return the image from 'key', if it does not
// exists a 'nil' will be returned
func (i *ImagesCache) Get(key string) *ebiten.Image {
	ei, _ := i.images[key]

	return ei
}

func init() {
	Images = &ImagesCache{
		images: make(map[string]*ebiten.Image),
	}

	for _, u := range unit.Units {
		Images.images[u.FacesetKey()] = ebiten.NewImageFromImage(u.Faceset)
		Images.images[u.WalkKey()] = ebiten.NewImageFromImage(u.Walk)
		Images.images[u.ProfileKey()] = ebiten.NewImageFromImage(u.Profile)
		if u.HasAbility(ability.Attack) {
			Images.images[u.AttackKey()] = ebiten.NewImageFromImage(u.Attack)
			Images.images[u.IdleKey()] = ebiten.NewImageFromImage(u.Idle)
		}
	}
	for k, i := range ability.Images {
		Images.images[ability.Key(k)] = ebiten.NewImageFromImage(i)
	}
	for _, t := range tower.Towers {
		Images.images[t.FacesetKey()] = ebiten.NewImageFromImage(t.Faceset)
		Images.images[t.IdleKey()] = ebiten.NewImageFromImage(t.Idle)
		Images.images[t.ProfileKey()] = ebiten.NewImageFromImage(t.Profile)
	}
	for i, m := range store.MapImages {
		Images.images[fmt.Sprintf(store.MapImageKeyFmt, i)] = ebiten.NewImageFromImage(m)
	}

	tli, _, err := image.Decode(bytes.NewReader(assets.TilesetLogic_png))
	if err != nil {
		panic(err)
	}

	ai, _, err := image.Decode(bytes.NewReader(assets.Arrow_png))
	if err != nil {
		panic(err)
	}

	Images.images[CrossImageKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tli).SubImage(image.Rect(4*16, 5*16, 4*16+16, 5*16+16)))
	Images.images[ArrowImageKey] = ebiten.NewImageFromImage(ai)

	tsn, _, err := image.Decode(bytes.NewReader(assets.TilesetNature_png))
	if err != nil {
		panic(err)
	}
	Images.images[BuffBurrowedReadyKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tsn).SubImage(image.Rect(4*16, 17*16, 4*16+16, 17*16+16)))
	Images.images[BuffBurrowedKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tsn).SubImage(image.Rect(6*16, 17*16, 6*16+16, 17*16+16)))

	Images.images[CapIconKey] = ebiten.NewImageFromImage(Images.images[unit.Units[unit.Ninja.String()].WalkKey()].SubImage(image.Rect(0, 0, 16, 16)))

	vli, _, err := image.Decode(bytes.NewReader(assets.VillagerIdle_png))
	if err != nil {
		panic(err)
	}
	Images.images[PlayerIconKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(vli).SubImage(image.Rect(0, 0, 16, 16)))

	hearts, _, err := image.Decode(bytes.NewReader(assets.Hearts_png))
	if err != nil {
		panic(err)
	}
	Images.images[HeartIconKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(hearts).SubImage(image.Rect(4*16, 0, 4*16+16, 16)))

	var keyImage = map[string][]byte{
		BuffResurrectingKey:          assets.SkeletonDemonDead_png,
		LifeBarProgressKey:           assets.LifeBarMiniProgress_png,
		LifeBarUnderKey:              assets.LifeBarMiniUnder_png,
		ShieldBarProgressKey:         assets.ShieldBarMiniProgress_png,
		LifeBarBigProgressKey:        assets.LifeBarBigProgress_png,
		LifeBarBigUnderKey:           assets.LifeBarBigUnder_png,
		BGKey:                        assets.BG_png,
		ButtonPressedKey:             assets.ButtonPressed_png,
		ButtonNormalKey:              assets.ButtonNormal_png,
		ButtonHoverKey:               assets.ButtonHover_png,
		ButtonDisabledKey:            assets.ButtonDisabled_png,
		BigButtonPressedKey:          assets.BigButtonPressed_png,
		BigButtonNormalKey:           assets.BigButtonNormal_png,
		BigButtonHoverKey:            assets.BigButtonHover_png,
		BigButtonDisabledKey:         assets.BigButtonDisabled_png,
		MenuButtonPressedKey:         assets.MenuButtonPressed_png,
		MenuButtonHoverKey:           assets.MenuButtonHover_png,
		LogoKey:                      assets.Logo_png,
		Border4Key:                   assets.Border4_png,
		SetupGameBGKey:               assets.SetupGameBG_png,
		SetupGameFrameKey:            assets.SetupGameFrame_png,
		BigButtonPressedLeftTabKey:   assets.BigButtonPressedLeftTab_png,
		BigButtonNormalLeftTabKey:    assets.BigButtonNormalLeftTab_png,
		BigButtonHoverLeftTabKey:     assets.BigButtonHoverLeftTab_png,
		BigButtonDisabledLeftTabKey:  assets.BigButtonDisabledLeftTab_png,
		BigButtonPressedRightTabKey:  assets.BigButtonPressedRightTab_png,
		BigButtonNormalRightTabKey:   assets.BigButtonNormalRightTab_png,
		BigButtonHoverRightTabKey:    assets.BigButtonHoverRightTab_png,
		BigButtonDisabledRightTabKey: assets.BigButtonDisabledRightTab_png,
		LobbiesTableKey:              assets.LobbiesTable_png,
		DarkInputBGKey:               assets.DarkInputBG_png,
		GrayInputBGKey:               assets.GrayInputBG_png,

		HSliderGrabberKey:         assets.HSliderGrabber_png,
		HSliderGrabberHoverKey:    assets.HSliderGrabberHover_png,
		HSliderGrabberDisabledKey: assets.HSliderGrabberDisabled_png,

		CheckboxCheckedKey:   assets.CheckboxChecked_png,
		CheckboxUncheckedKey: assets.CheckboxUnchecked_png,

		BigCancelButtonNormalKey:  assets.BigCancelButtonNormal_png,
		BigCancelButtonHoverKey:   assets.BigCancelButtonHover_png,
		BigCancelButtonPressedKey: assets.BigCancelButtonPressed_png,

		DisplayDefaultBGKey:       assets.DisplayDefaultBG_png,
		DisplayDefaultTowersBGKey: assets.DisplayDefaultTowersBG_png,

		ButtonBorderNormalKey:   assets.ButtonBorderNormal_png,
		ButtonBorderPressedKey:  assets.ButtonBorderPressed_png,
		ButtonBorderHoverKey:    assets.ButtonBorderHover_png,
		ButtonBorderDisabledKey: assets.ButtonBorderDisabled_png,

		ScoreboardRowBGKey:        assets.ScoreboardRowBG_png,
		ScoreboardRowCurrentBGKey: assets.ScoreboardRowCurrentBG_png,

		GoldIconKey:        assets.GoldIcon_png,
		IncomeIconKey:      assets.IncomeIcon_png,
		IncomeTimerIconKey: assets.IncomeTimerIcon_png,

		InfoLeftBGKey:   assets.InfoLeftBG_png,
		InfoMiddleBGKey: assets.InfoMiddleBG_png,
		InfoRightBGKey:  assets.InfoRightBG_png,

		UnitUpdateButtonAnimationKey: assets.UnitUpdateButtonAnimation_png,

		SellIconKey: assets.SellIcon_png,

		DisplayTargetImageBGKey:   assets.DisplayTargetImageBG,
		DisplayTargetDetailsBGKey: assets.DisplayTargetDetailsBG,

		DamageIconKey:        assets.DamageIcon_png,
		AttackSpeedIconKey:   assets.AttackSpeedIcon_png,
		RangeIconKey:         assets.RangeIcon_png,
		MovementSpeedIconKey: assets.MovementSpeedIcon_png,

		ToolTipBGKey: assets.ToolTipBG_png,
	}

	for k, b := range keyImage {
		i, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			log.Fatalf("failed to decode image with key %q: %s", k, err)
		}
		Images.images[k] = ebiten.NewImageFromImage(i)
	}

}
