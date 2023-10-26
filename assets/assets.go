package assets

import (
	_ "embed"
	_ "image/png"
)

//go:embed cyclope/Faceset.png
var CyclopeFaceset_png []byte

//go:embed TilesetHouse.png
var TilesetHouse_png []byte

//go:embed TilesetLogic.png
var TilesetLogic_png []byte

//go:embed TilesetElement.png
var TilesetElement_png []byte

//go:embed cyclope/Cyclopes.png
var Cyclopes_png []byte

//go:embed maps/2.png
var Map_2_png []byte

//go:embed maps/3.png
var Map_3_png []byte

//go:embed maps/4.png
var Map_4_png []byte

//go:embed maps/5.png
var Map_5_png []byte

//go:embed maps/6.png
var Map_6_png []byte

//go:embed YesButton.png
var YesButton_png []byte

//go:embed NoButton.png
var NoButton_png []byte

//go:embed towers.json
var Towers_json []byte

//go:embed units.json
var Units_json []byte

//go:embed NormalFont.ttf
var NormalFont_ttf []byte
