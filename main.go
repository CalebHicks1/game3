package main

import (
	"image"
	"math"
	"math/rand/v2"
	"os"
	"time"

	_ "image/png"

	_ "net/http/pprof"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)

// CONSTS ///////////////////////////////////////////////////////////////////////////////////////////////
const (
	gridWidth  = 300
	gridHeight = 150
	tileSize   = 32
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func initGrid(grid [gridWidth][gridHeight]*Tile) [gridWidth][gridHeight]*Tile {
	// build the grid, randomly assign floor and wall tiles
	for x := 0; x < gridWidth; x++ {
		for y := 0; y < gridHeight; y++ {
			tile := Tile{
				X:    x,
				Y:    y,
				Type: TYPE_FLOOR,
			}
			if rand.Float64() < 0.38 {
				tile.Type = TYPE_WALL
			}
			grid[x][y] = &tile
		}
	}
	return grid
}

func run() {

	// SETUP ////////////////////////////////////////////////////////////////////////////////////////////////

	// window config
	cfg := opengl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1200, 800),
		VSync:  true,
	}
	// create new window
	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// load tree spritesheet
	spritesheet, err := loadPicture("trees.png")
	if err != nil {
		panic(err)
	}

	// create tree frames from spritesheet
	var treesFrames []pixel.Rect
	for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += 32 {
		for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += 32 {
			treesFrames = append(treesFrames, pixel.R(x, y, x+32, y+32))
		}
	}

	// VARS /////////////////////////////////////////////////////////////////////////////////////////////////
	var (
		camPos       = pixel.ZV
		camSpeed     = 500.0
		camZoom      = 1.0
		camZoomSpeed = 1.2
		// trees        []*pixel.Sprite
		// matrices     []pixel.Matrix
		tileGrid [gridWidth][gridHeight]*Tile
		last     = time.Now()
		imd      = imdraw.New(nil)
	)

	// INIT /////////////////////////////////////////////////////////////////////////////////////////////////
	tileGrid = initGrid(tileGrid)

	// GAME LOOP /////////////////////////////////////////////////////////////////////////////////////////////
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(pixel.ZV.Sub(camPos))
		// .Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		// if win.JustPressed(pixel.MouseButtonLeft) {
		// 	tree := pixel.NewSprite(spritesheet, treesFrames[rand.Intn(len(treesFrames))])
		// 	trees = append(trees, tree)
		// 	mouse := cam.Unproject(win.MousePosition())
		// 	matrices = append(matrices, pixel.IM.Scaled(pixel.ZV, 4).Moved(mouse))
		// }
		if win.Pressed(pixel.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixel.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixel.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixel.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		win.Clear(colornames.Forestgreen)
		imd.Clear()

		// for i, tree := range trees {
		// 	tree.Draw(win, matrices[i])
		// }

		// draw the grid
		for x := 0; x < gridWidth; x++ {
			for y := 0; y < gridHeight; y++ {
				tile := tileGrid[x][y]
				if tile.Type == TYPE_WALL {
					imd.Color = pixel.RGB(0.1, 0.1, 0.1)
				} else {
					imd.Color = pixel.RGB(0.1, 0.2, 0.1)
				}
				// imd.Color = pixel.RGB(0, 0, 0)
				// draw bottom left of tile
				imd.Push(pixel.V(float64(x*tileSize), float64(y*tileSize)))

				// draw top right of tile
				imd.Push(pixel.V(float64((x+1)*tileSize), float64((y+1)*tileSize)))
				imd.Rectangle(0)
			}
		}

		imd.Draw(win)
		win.Update()

		if win.JustPressed(pixel.KeySpace) {
			// do a round of cellular automata
			// create a new grid to store the results
			var newGrid [gridWidth][gridHeight]*Tile
			for x := 0; x < gridWidth; x++ {
				for y := 0; y < gridHeight; y++ {
					newTile := Tile{
						X:    x,
						Y:    y,
						Type: TYPE_FLOOR,
					}
					newGrid[x][y] = &newTile
					// get neighbor nodes
					wallCount := 0
					floorCount := 0
					for dx := -1; dx <= 1; dx++ {
						for dy := -1; dy <= 1; dy++ {
							// skip the center node
							if dx == 0 && dy == 0 {
								continue
							}
							// check bounds
							if x+dx < 0 || x+dx >= gridWidth || y+dy < 0 || y+dy >= gridHeight {
								continue
							}
							// check neighbor type
							if tileGrid[x+dx][y+dy].Type == TYPE_WALL {
								wallCount++
							} else {
								floorCount++
							}
						}
					}
					if wallCount >= 4 {
						newTile.Type = TYPE_WALL
					}
				}
			}
			// copy the new grid to the old grid
			tileGrid = newGrid
		}
		if win.JustPressed(pixel.KeyR) {
			tileGrid = initGrid(tileGrid)
		}
	}
}

func main() {
	opengl.Run(run)
}
