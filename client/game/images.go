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
)

var (
	imagesCache *ImagesCache
)

const (
	crossImageKey string = "cross-image"
	arrowImageKey string = "arrow"
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
		imagesCache.images[u.SpriteKey()] = ebiten.NewImageFromImage(u.Sprite)
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
}

// Get will return the image from 'key', if it does not
// exists a 'nil' will be returned
func (i *ImagesCache) Get(key string) *ebiten.Image {
	ei, _ := i.images[key]

	return ei
}
