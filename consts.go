package main

// create a tile type enum
type TileType int

const (
	// tile types
	TYPE_WALL  = 0
	TYPE_FLOOR = 1

	// grid dimensions
	gridWidth  = 300
	gridHeight = 150
	tileSize   = 32
)

type Tile struct {
	// X and Y are the coordinates of the tile in the map.
	X, Y int
	// Type is the type of the tile.
	Type TileType
}

type Corner struct {
	X, Y float64
}

type Player struct {
	// X and Y are the coordinates of the player in the map.
	X, Y      float64
	walkSpeed float64
	runSpeed  float64
}
