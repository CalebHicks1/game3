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

func iterateGrid(grid [gridWidth][gridHeight]*Tile) [gridWidth][gridHeight]*Tile {
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
					if grid[x+dx][y+dy].Type == TYPE_WALL {
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
	return newGrid
}

func drawGrid(grid [gridWidth][gridHeight]*Tile, batch *pixel.Batch, spritesheet pixel.Picture, spriteFrames []pixel.Rect) {
	// redraw the grid
	batch.Clear()
	for x := 0; x < gridWidth; x++ {
		for y := 0; y < gridHeight; y++ {
			tile := grid[x][y]
			if tile.Type == TYPE_WALL {

				// determine which sprite to use
				// get neighbor nodes
				hasLeftNeighbor := false
				hasRightNeighbor := false
				hasTopNeighbor := false
				hasBottomNeighbor := false
				if x > 0 && grid[x-1][y].Type == TYPE_WALL {
					hasLeftNeighbor = true
				}
				if x < gridWidth-1 && grid[x+1][y].Type == TYPE_WALL {
					hasRightNeighbor = true
				}
				if y > 0 && grid[x][y-1].Type == TYPE_WALL {
					hasBottomNeighbor = true
				}
				if y < gridHeight-1 && grid[x][y+1].Type == TYPE_WALL {
					hasTopNeighbor = true
				}

				frameNum := 0
				// determine which sprite to use
				if hasLeftNeighbor && hasRightNeighbor && hasTopNeighbor && hasBottomNeighbor {
					// all neighbors
					frameNum = 3
				}
				if hasLeftNeighbor && hasRightNeighbor && !hasTopNeighbor && !hasBottomNeighbor {
					// left right,  neighbors
					frameNum = 2
				}
				if !hasLeftNeighbor && hasRightNeighbor && !hasTopNeighbor && hasBottomNeighbor {
					// right bottom neighbors
					frameNum = 1
				}
				if hasLeftNeighbor && !hasRightNeighbor && !hasTopNeighbor && hasBottomNeighbor {
					// left bottom neighbors
					frameNum = 5
				}
				if hasLeftNeighbor && hasRightNeighbor && !hasTopNeighbor && hasBottomNeighbor {
					// left bottom right  neighbors
					frameNum = 4
				}
				if hasLeftNeighbor && !hasRightNeighbor && hasTopNeighbor && hasBottomNeighbor {
					// left bottom top  neighbors
					frameNum = 6
				}
				if !hasLeftNeighbor && hasRightNeighbor && hasTopNeighbor && hasBottomNeighbor {
					// left bottom top  neighbors
					frameNum = 7
				}
				if hasLeftNeighbor && hasRightNeighbor && hasTopNeighbor && !hasBottomNeighbor {
					// left bottom top  neighbors
					frameNum = 8
				}
				if !hasLeftNeighbor && !hasRightNeighbor && !hasTopNeighbor && hasBottomNeighbor {
					// bottom neighbors
					frameNum = 9
				}
				if hasLeftNeighbor && !hasRightNeighbor && hasTopNeighbor && !hasBottomNeighbor {
					// left top neighbors
					frameNum = 10
				}
				if hasLeftNeighbor && !hasRightNeighbor && !hasTopNeighbor && !hasBottomNeighbor {
					// left neighbors
					frameNum = 11
				}
				if !hasLeftNeighbor && hasRightNeighbor && hasTopNeighbor && !hasBottomNeighbor {
					// right top neighbors
					frameNum = 12
				}
				if !hasLeftNeighbor && hasRightNeighbor && !hasTopNeighbor && !hasBottomNeighbor {
					// right neighbors
					frameNum = 13
				}
				if !hasLeftNeighbor && !hasRightNeighbor && hasTopNeighbor && !hasBottomNeighbor {
					// top neighbors
					frameNum = 15
				}
				if !hasLeftNeighbor && !hasRightNeighbor && hasTopNeighbor && hasBottomNeighbor {
					// top neighbors
					frameNum = 14
				}

				wallSprite := pixel.NewSprite(spritesheet, spriteFrames[frameNum])
				wallSprite.Draw(batch, pixel.IM.Moved(pixel.V(float64(x*tileSize), float64(y*tileSize))))
			}
			// imd.Color = pixel.RGB(0, 0, 0)
			// draw bottom left of tile

			// // draw top right of tile
			// imd.Rectangle(0)
		}
	}
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

	// draw the player as a 1x2 red rectangle
	canvas := opengl.NewCanvas(pixel.R(0, 0, 1200, 800))
	lightCanvas := opengl.NewCanvas(pixel.R(0, 0, 1200, 800))
	shadowCanvas := opengl.NewCanvas(pixel.R(0, 0, 1200, 800))
	// load tree spritesheet
	spritesheet, err := loadPicture("sprites/wall.png")
	if err != nil {
		panic(err)
	}

	// create tree frames from spritesheet
	var wallFrames []pixel.Rect
	for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += 32 {
		for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += 32 {
			wallFrames = append(wallFrames, pixel.R(x, y, x+32, y+32))
		}
	}

	// VARS /////////////////////////////////////////////////////////////////////////////////////////////////
	var (
		camPos = pixel.V(1300, 1300)
		//camSpeed     = 500.0
		camZoom      = 1.0
		camZoomSpeed = 1.2
		// trees        []*pixel.Sprite
		// matrices     []pixel.Matrix
		tileGrid  [gridWidth][gridHeight]*Tile
		last      = time.Now()
		imd       = imdraw.New(nil)
		wallBatch = pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)
	)

	// INIT /////////////////////////////////////////////////////////////////////////////////////////////////
	tileGrid = initGrid(tileGrid)

	// iterate the cellular automata a few times
	for x := 0; x < 7; x++ {
		tileGrid = iterateGrid(tileGrid)
	}
	drawGrid(tileGrid, wallBatch, spritesheet, wallFrames)

	// init player
	player := Player{
		X:         1300,
		Y:         1300,
		walkSpeed: 200.0,
		runSpeed:  400.0,
	}

	// GAME LOOP /////////////////////////////////////////////////////////////////////////////////////////////

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// .Moved(win.Bounds().Center().Sub(camPos))

		// if win.JustPressed(pixel.MouseButtonLeft) {
		// 	tree := pixel.NewSprite(spritesheet, treesFrames[rand.Intn(len(treesFrames))])
		// 	trees = append(trees, tree)
		// 	mouse := cam.Unproject(win.MousePosition())
		// 	matrices = append(matrices, pixel.IM.Scaled(pixel.ZV, 4).Moved(mouse))
		// }

		// camera movements
		// interpolate camera to follow player
		// camPos = pixel.Lerp(camPos, pixel.V(player.X, player.Y), 1-math.Pow(1.0/128, dt))

		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)
		// // cam := pixel.IM.Moved(camPos.Scaled(-1)).Moved(camPos.Sub(win.Bounds().Center()))
		// cam := pixel.IM.Scaled(camPos, camZoom).Moved(pixel.ZV.Sub(camPos))

		// lerp the camera position towards the gopher
		camPos = pixel.Lerp(camPos, pixel.V(player.X, player.Y), 1-math.Pow(1.0/128, dt))
		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))

		canvas.SetMatrix(cam)
		// lightCanvas.SetMatrix(cam)
		// shadowCanvas.SetMatrix(cam)
		// win.SetMatrix(cam)

		// player movements
		currSpeed := player.walkSpeed
		if win.Pressed(pixel.KeyLeftShift) {
			currSpeed = player.runSpeed
		}
		if win.Pressed(pixel.KeyA) {
			player.X -= currSpeed * dt
		}
		if win.Pressed(pixel.KeyD) {
			player.X += currSpeed * dt
		}
		if win.Pressed(pixel.KeyS) {
			player.Y -= currSpeed * dt
		}
		if win.Pressed(pixel.KeyW) {
			player.Y += currSpeed * dt
		}

		// draw the grid

		imd.Clear()
		if win.JustPressed(pixel.KeySpace) {
			// iterate the grid
			tileGrid = iterateGrid(tileGrid)
			// redraw the grid
			wallBatch.Clear()
			drawGrid(tileGrid, wallBatch, spritesheet, wallFrames)

		}
		if win.JustPressed(pixel.KeyR) {
			tileGrid = initGrid(tileGrid)
			// redraw the grid
			wallBatch.Clear()
			drawGrid(tileGrid, wallBatch, spritesheet, wallFrames)
		}

		// draw the grid to the canvas
		//wallBatch.Draw(canvas)

		// draw lights
		// imd.Clear()
		// imd.Color = pixel.RGB(0, 1, 1)
		// imd.Push(pixel.V(float64(1300), float64(1300)))
		// // imd.Push(pixel.V(float64(1350), float64(1350)))
		// imd.Circle(50, 0)
		// imd.Draw(lightCanvas)

		// lightSprite := pixel.NewSprite(lightCanvas, lightCanvas.Bounds())
		// lightSprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		//lightTexture = lightCanvas.Texture().
		// win.SetColorMask(pixel.Alpha(1))

		// lightCanvas.SetColorMask(pixel.RGB(1, 1, 1))
		// lightCanvas.SetColorMask(pixel.Alpha(1))
		// win.SetColorMask(pixel.RGB(1, 0, 0))

		// 1 draw everything inside the light
		// 2 draw everything outside the light

		// win.SetColorMask(pixel.RGB(1, 0, 0))
		//sprite.Draw(lightCanvas, pixel.IM.Moved(win.Bounds().Center()))
		//canvas.SetColorMask(pixel.RGB(1, 1, 1))

		canvas.Clear(pixel.RGB(1, 1, 1))
		// draw the world to the canvas
		// create a background rectangle
		imd.Color = pixel.RGB(0.154, 0.139, 0.152)
		imd.Push(pixel.V(0, 0))
		imd.Push(pixel.V(1200*tileSize, 800*tileSize))
		imd.Rectangle(0)
		imd.Draw(canvas)
		imd.Clear()
		// draw the grid to the canvas
		wallBatch.Draw(canvas)
		// draw the player to the canvas
		imd.Color = pixel.RGB(0.8, 0.2, 0.2)
		imd.Push(pixel.V(float64(player.X), float64(player.Y)))
		imd.Push(pixel.V(float64(player.X+tileSize), float64(player.Y+(2*tileSize))))
		imd.Rectangle(0)
		imd.Draw(canvas)

		// create a sprite for the light gradient
		lightPic, err := loadPicture("sprites/light.png")
		if err != nil {
			panic(err)
		}
		lightSprite := pixel.NewSprite(lightPic, lightPic.Bounds())

		mousePos := win.MousePosition()

		// draw the light and shadow canvases

		// lightCanvas should be everything inside the light
		lightCanvas.Clear(pixel.Alpha(0))
		lightCanvas.SetComposeMethod(pixel.ComposeOver)
		lightSprite.Draw(lightCanvas, pixel.IM.Scaled(pixel.ZV, 2).Moved(pixel.V(mousePos.X, mousePos.Y)))
		lightCanvas.SetComposeMethod(pixel.ComposeIn)
		canvas.Draw(lightCanvas, pixel.IM.Moved(lightCanvas.Bounds().Center()))
		lightCanvas.SetColorMask(pixel.RGB(0.7, 0.7, 1))

		// shadowCanvas should be everything outside the light
		shadowCanvas.Clear(pixel.Alpha(0))
		shadowCanvas.SetComposeMethod(pixel.ComposeOver)
		lightSprite.Draw(shadowCanvas, pixel.IM.Scaled(pixel.ZV, 2).Moved(pixel.V(mousePos.X, mousePos.Y)))
		shadowCanvas.SetComposeMethod(pixel.ComposeOut)
		canvas.Draw(shadowCanvas, pixel.IM.Moved(shadowCanvas.Bounds().Center()))
		shadowCanvas.SetColorMask(pixel.RGB(0.2, 0.2, 0.5))

		// draw the light and shadow canvases to the window
		win.Clear(pixel.RGB(0, 1, 0))
		win.SetComposeMethod(pixel.ComposeOver)
		// draw the canvas again so that the background color doesn't bleed through
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		lightCanvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		shadowCanvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()
	}
}

func main() {
	opengl.Run(run)
}
