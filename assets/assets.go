package assets

import (
	_ "embed"
	_ "image/png"
)

//go:embed cyclope/Faceset.png
var CyclopeFaceset_png []byte

//go:embed TilesetHouse.png
var TilesetHouse_png []byte

//go:embed cyclope/Cyclopes.png
var Cyclopes_png []byte

//go:embed maps/1v1.png
var M_1v1_png []byte

//go:embed towers.json
var Towers_json []byte

//go:embed units.json
var Units_json []byte
