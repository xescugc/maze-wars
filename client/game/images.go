package game

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
)

var (
	imagesCache *ImagesCache
)

const (
	crossImageKey = "cross-image"
	arrowImageKey = "arrow"

	buffBurrowedKey      = "buff-burrowed"
	buffBurrowedReadyKey = "buff-burrowed-ready"

	buffResurrectingKey = "buff-resurrecting"

	lifeBarProgressKey   = "life-bar-progress"
	lifeBarUnderKey      = "life-bar-under"
	shieldBarProgressKey = "shield-bar-progress"

	lifeBarBigProgressKey = "life-bar-big-progress"
	lifeBarBigUnderKey    = "life-bar-big-under"
)

// ImagesCache is a simple cache for all the images, so instead
// of running 'ebiten.NewImageFromImage' we just ran it once
// and reuse it all the time
type ImagesCache struct {
	images map[string]*ebiten.Image
}

func init() {
	imagesCache = &ImagesCache{
		images: make(map[string]*ebiten.Image),
	}

	for _, u := range unit.Units {
		imagesCache.images[u.FacesetKey()] = ebiten.NewImageFromImage(u.Faceset)
		imagesCache.images[u.WalkKey()] = ebiten.NewImageFromImage(u.Walk)
		if u.HasAbility(ability.Attack) {
			imagesCache.images[u.AttackKey()] = ebiten.NewImageFromImage(u.Attack)
			imagesCache.images[u.IdleKey()] = ebiten.NewImageFromImage(u.Idle)
		}
	}
	for _, t := range tower.Towers {
		imagesCache.images[t.FacesetKey()] = ebiten.NewImageFromImage(t.Faceset)
	}
	for i, m := range store.MapImages {
		imagesCache.images[fmt.Sprintf(store.MapImageKeyFmt, i)] = ebiten.NewImageFromImage(m)
	}

	tli, _, err := image.Decode(bytes.NewReader(assets.TilesetLogic_png))
	if err != nil {
		panic(err)
	}

	ai, _, err := image.Decode(bytes.NewReader(assets.Arrow_png))
	if err != nil {
		panic(err)
	}

	imagesCache.images[crossImageKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tli).SubImage(image.Rect(4*16, 5*16, 4*16+16, 5*16+16)))
	imagesCache.images[arrowImageKey] = ebiten.NewImageFromImage(ai)

	tsn, _, err := image.Decode(bytes.NewReader(assets.TilesetNature_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[buffBurrowedReadyKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tsn).SubImage(image.Rect(4*16, 17*16, 4*16+16, 17*16+16)))
	imagesCache.images[buffBurrowedKey] = ebiten.NewImageFromImage(ebiten.NewImageFromImage(tsn).SubImage(image.Rect(6*16, 17*16, 6*16+16, 17*16+16)))

	sdd, _, err := image.Decode(bytes.NewReader(assets.SkeletonDemonDead_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[buffResurrectingKey] = ebiten.NewImageFromImage(sdd)

	lbpi, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniProgress_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[lifeBarProgressKey] = ebiten.NewImageFromImage(lbpi)

	lbui, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniUnder_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[lifeBarUnderKey] = ebiten.NewImageFromImage(lbui)

	sbpi, _, err := image.Decode(bytes.NewReader(assets.ShieldBarMiniProgress_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[shieldBarProgressKey] = ebiten.NewImageFromImage(sbpi)

	lbbpi, _, err := image.Decode(bytes.NewReader(assets.LifeBarBigProgress_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[lifeBarBigProgressKey] = ebiten.NewImageFromImage(lbbpi)

	lbbui, _, err := image.Decode(bytes.NewReader(assets.LifeBarBigUnder_png))
	if err != nil {
		panic(err)
	}
	imagesCache.images[lifeBarBigUnderKey] = ebiten.NewImageFromImage(lbbui)
}

// Get will return the image from 'key', if it does not
// exists a 'nil' will be returned
func (i *ImagesCache) Get(key string) *ebiten.Image {
	ei, _ := i.images[key]

	return ei
}
