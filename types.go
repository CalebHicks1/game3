package main

// create a tile type enum
type TileType int

const (
	Wall TileType = iota
	Floor
)

type Tile struct {
	// X and Y are the coordinates of the tile in the map.
	X, Y int
	// Type is the type of the tile.
	Type TileType
}
