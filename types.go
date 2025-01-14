package main

// create a tile type enum
type TileType int

const (
	TYPE_WALL  = 0
	TYPE_FLOOR = 1
)

type Tile struct {
	// X and Y are the coordinates of the tile in the map.
	X, Y int
	// Type is the type of the tile.
	Type TileType
}
