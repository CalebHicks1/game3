package main

import (
	"image"
	"math"
	"math/rand"
	"os"
	"time"

	_ "image/png"

	_ "net/http/pprof"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
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

func run() {

	// SETUP ////////////////////////////////////////////////////////////////////////////////////////////////

	// window config
	cfg := opengl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
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

	// CONSTS ///////////////////////////////////////////////////////////////////////////////////////////////
	const (
		gridWidth  = 10
		gridHeight = 10
		tileSize   = 32
	)

	// VARS /////////////////////////////////////////////////////////////////////////////////////////////////
	var (
		camPos       = pixel.ZV
		camSpeed     = 500.0
		camZoom      = 1.0
		camZoomSpeed = 1.2
		trees        []*pixel.Sprite
		matrices     []pixel.Matrix
		tileGrid     [gridWidth][gridHeight]*Tile
		last         = time.Now()
		imd          = imdraw.New(nil)
	)

	// INIT /////////////////////////////////////////////////////////////////////////////////////////////////
	// build the grid, randomly assign floor and wall tiles
	for x := 0; x < gridWidth; x++ {
		for y := 0; y < gridHeight; y++ {
			tile := Tile{
				X:    x,
				Y:    y,
				Type: Floor,
			}
			if rand.Float64() < 0.3 {
				tile.Type = Wall
			}
			tileGrid[x][y] = &tile
		}
	}

	// GAME LOOP /////////////////////////////////////////////////////////////////////////////////////////////
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixel.MouseButtonLeft) {
			tree := pixel.NewSprite(spritesheet, treesFrames[rand.Intn(len(treesFrames))])
			trees = append(trees, tree)
			mouse := cam.Unproject(win.MousePosition())
			matrices = append(matrices, pixel.IM.Scaled(pixel.ZV, 4).Moved(mouse))
		}
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

		for i, tree := range trees {
			tree.Draw(win, matrices[i])
		}

		// draw the grid
		for x := 0; x < gridWidth; x++ {
			for y := 0; y < gridHeight; y++ {
				imd.Color = pixel.RGB(0, 0, 0)
				// draw bottom left of tile
				imd.Push(pixel.V(float64(x*tileSize), float64(y*tileSize)))
				// draw bottom right of tile
				imd.Push(pixel.V(float64((x+1)*tileSize), float64(y*tileSize)))
				// draw top left of tile
				imd.Push(pixel.V(float64(x*tileSize), float64((y+1)*tileSize)))
				// draw top right of tile
				imd.Push(pixel.V(float64((x+1)*tileSize), float64((y+1)*tileSize)))
				imd.Rectangle(0)
			}
		}

		imd.Draw(win)
		win.Update()
	}
}

func main() {
	opengl.Run(run)
}
