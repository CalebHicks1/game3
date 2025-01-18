package main

import (
	"image"
	"math"
	"math/rand/v2"
	"os"
	"time"

	_ "image/png"

	_ "net/http/pprof"

	"github.com/go-gl/mathgl/mgl32"
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

				// sprites are anchored at the center, so move them to the bottom left of the tile
				wallSprite := pixel.NewSprite(spritesheet, spriteFrames[frameNum])
				wallSprite.Draw(batch, pixel.IM.Moved(pixel.V(float64((x*tileSize)-(tileSize/2)), float64((y*tileSize)-(tileSize/2)))))
			}
			// imd.Color = pixel.RGB(0, 0, 0)
			// draw bottom left of tile

			// // draw top right of tile
			// imd.Rectangle(0)
		}
	}
}

// func drawToCanvas(canvas *opengl.Canvas, grid [gridWidth][gridHeight]*Tile, imd *imdraw.IMDraw) {
// 	imd.Clear()
// 	for x := 0; x < gridWidth; x++ {
// 		for y := 0; y < gridHeight; y++ {
// 			tile := grid[x][y]
// 			imd.Push(pixel.V(float64(tile.X*tileSize), float64(tile.Y*tileSize)))
// 			imd.Circle(2, 0)
// 		}
// 	}
// }

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
	shadowCanvas := opengl.NewCanvas(pixel.R(0, 0, 1200, 800))
	debugCanvas := opengl.NewCanvas(pixel.R(0, 0, 1200, 800))

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
		tileGrid     [gridWidth][gridHeight]*Tile
		last         = time.Now()
		imd          = imdraw.New(nil)
		wallBatch    = pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)
		enableLights = false
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

	// SHADER ///////////////////////////////////////////////////////////////////////////////////////////////
	// create uniform with light position
	// need to convert the player world coords to screen coords

	lightPos := mgl32.Vec2{0, 0}
	// aspectRatio := float32(win.Bounds().W() / win.Bounds().H())
	win.Canvas().SetUniform("uLightPos", &lightPos)
	// win.Canvas().SetUniform("uAspectRatio", &aspectRatio)

	var fragmentShader = `
			#version 330 core

			in vec2  vTexCoords;

			out vec4 fragColor;

			uniform vec4 uTexBounds;
			uniform sampler2D uTexture;
			uniform vec2 uLightPos;
			// Function to generate a small random noise value
			float rand(vec2 co){
				return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
			}

			void main() {
				// Get our current screen coordinate
				vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;

				// calculate the distance from the light
				float dist = distance(t, uLightPos);

				// calculate the light intensity
				float intensity = 1 / (1.0 + dist*5);

				// add a small amount of noise to the intensity
				intensity += rand(t) * 0.03;

				// get the color from the texture
				vec4 color = texture(uTexture, t);

				
				vec3 spotLightColor = vec3(1, 0, 0);
				vec3 ambientColor = vec3(0.23, 0.23, 0.38);
				color.rgb = (color.rgb * ambientColor) + (color.rgb * intensity * spotLightColor);
				fragColor = color;
			}
		`

	// GAME LOOP /////////////////////////////////////////////////////////////////////////////////////////////
	win.Canvas().SetFragmentShader(fragmentShader)

	// firstLoop := true
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
		shadowCanvas.SetMatrix(cam)
		debugCanvas.SetMatrix(cam)

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

		// clear all the canvases
		win.Clear(pixel.RGB(0, 0, 0))
		canvas.Clear(pixel.RGB(0.154, 0.139, 0.152))
		shadowCanvas.Clear(pixel.Alpha(0.5))
		debugCanvas.Clear(pixel.Alpha(0))
		imd.Clear()

		// draw the grid to the canvas
		wallBatch.Draw(canvas)

		// draw the player to the canvas
		imd.Color = pixel.RGB(1, 1, 1)
		imd.Push(pixel.V(float64(player.X), float64(player.Y)))
		imd.Push(pixel.V(float64(player.X+tileSize), float64(player.Y+(2*tileSize))))
		imd.Rectangle(0)
		imd.Draw(canvas)

		// create a texture from the canvas and draw it to the window

		mousePos := cam.Unproject(win.MousePosition())
		// currently, we have the player pos in game space. Need to convert to screen space
		// to pass to the shader
		// TODO: create function to translate from game to screen coords
		playerScreenPos := cam.Project(pixel.V(mousePos.X, mousePos.Y))

		lightPos = mgl32.Vec2{float32(playerScreenPos.X / 1200), float32(playerScreenPos.Y / 800)}
		if enableLights {
			imd.Clear()
			imd.Color = pixel.RGB(1, 1, 1)
			imd.Push(pixel.V(mousePos.X-100, mousePos.Y-100))
			imd.Push(pixel.V(float64(mousePos.X+100), float64(mousePos.Y+100)))
			imd.Rectangle(0)
			imd.Draw(shadowCanvas)
			shadowSprite := pixel.NewSprite(shadowCanvas, canvas.Bounds())
			shadowSprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

			win.SetComposeMethod(pixel.ComposeIn)
		}

		sprite := pixel.NewSprite(canvas, canvas.Bounds())
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		// debug grid
		// imd.Color = pixel.RGB(1, 0, 0)
		// if firstLoop {
		// 	imd.Clear()
		// 	for x := 0; x < gridWidth; x++ {
		// 		for y := 0; y < gridHeight; y++ {
		// 			tile := tileGrid[x][y]
		// 			imd.Push(pixel.V(float64(tile.X*tileSize), float64(tile.Y*tileSize)))
		// 			imd.Circle(2, 0)
		// 		}
		// 	}
		// 	imd.Draw(debugCanvas)
		// 	firstLoop = false
		// }

		// calculate shadows with a tile based method
		// draw rays out to the corners of the nearest tiles.
		// 1. calculate corners of tiles nearby.
		// 2. draw rays to each corner, continue until we hit a wall.
		// 3. fill in the polygon with a light color.

		// step 1: calculate corners of tiles nearby
		// get the tile the mouse is on
		imd.Color = pixel.RGB(1, 0, 0)
		mouseTileX := int(mousePos.X / tileSize)
		mouseTileY := int(mousePos.Y / tileSize)
		// get the corners of the tile
		// top left
		topLeft := pixel.V(float64(mouseTileX*tileSize), float64(mouseTileY*tileSize))
		imd.Clear()
		imd.Push(topLeft)
		imd.Circle(2, 0)
		// top right
		topRight := pixel.V(float64((mouseTileX+1)*tileSize), float64(mouseTileY*tileSize))
		imd.Push(topRight)
		imd.Circle(2, 0)
		// bottom left
		bottomLeft := pixel.V(float64(mouseTileX*tileSize), float64((mouseTileY+1)*tileSize))
		imd.Push(bottomLeft)
		imd.Circle(2, 0)
		// bottom right
		bottomRight := pixel.V(float64((mouseTileX+1)*tileSize), float64((mouseTileY+1)*tileSize))
		imd.Push(bottomRight)
		imd.Circle(2, 0)

		// only iterate over nearby tiles
		tile_range := 10
		//var corner_points []int

		imd.Color = pixel.RGB(0, 0, 1)
		for x := mouseTileX - tile_range; x <= mouseTileX+tile_range; x++ {
			for y := mouseTileY - tile_range; y <= mouseTileY+tile_range; y++ {
				// check bounds
				if x < 0 || x >= gridWidth || y < 0 || y >= gridHeight {
					continue
				}
				// get the corners of the tile
				// top left
				topLeft := pixel.V(float64(x*tileSize), float64(y*tileSize))
				imd.Push(topLeft)
				imd.Circle(2, 0)
				// top right
				topRight := pixel.V(float64((x+1)*tileSize), float64(y*tileSize))
				imd.Push(topRight)
				imd.Circle(2, 0)
				// bottom left
				bottomLeft := pixel.V(float64(x*tileSize), float64((y+1)*tileSize))
				imd.Push(bottomLeft)
				imd.Circle(2, 0)
				// bottom right
				bottomRight := pixel.V(float64((x+1)*tileSize), float64((y+1)*tileSize))
				imd.Push(bottomRight)
				imd.Circle(2, 0)
			}
		}

		imd.Draw(debugCanvas)
		// win.SetComposeMethod()
		// create a texture from the canvas and draw it to the window

		debugCanvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()
	}
}

func main() {
	opengl.Run(run)
}
