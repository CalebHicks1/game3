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
		camPos = pixel.V(1300, 1300)
		//camSpeed     = 500.0
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

	// iterate the cellular automata a few times
	for x := 0; x < 4; x++ {
		tileGrid = iterateGrid(tileGrid)
	}

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
	win.Canvas().SetUniform("uLightPos", &lightPos)

	var fragmentShader = `
			#version 330 core

			in vec2  vTexCoords;

			out vec4 fragColor;

			uniform vec4 uTexBounds;
			uniform sampler2D uTexture;
			uniform vec2 uLightPos;

			void main() {
				// Get our current screen coordinate
				vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;

				// calculate the distance from the light
				float dist = distance(t, uLightPos);

				// calculate the light intensity
				float intensity = 0.5 / (1.0 + dist * dist);

				// get the color from the texture
				vec4 color = texture(uTexture, t);

				// apply the light intensity
				color *= intensity;
				
				fragColor = color;
			}
		`

	// GAME LOOP /////////////////////////////////////////////////////////////////////////////////////////////
	win.Canvas().SetFragmentShader(fragmentShader)
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

		win.Clear(pixel.RGB(0, 0, 0))
		canvas.Clear(pixel.RGB(0, 0, 0))
		imd.Clear()

		// convert player world coords to screen coords
		// playerScreenPos := cam.Unproject(pixel.V(player.X, player.Y))

		// xcoord := float32(playerScreenPos.X)
		// ycoord := float32(playerScreenPos.Y)
		lightPos = mgl32.Vec2{0, 0}

		// for i, tree := range trees {
		// 	tree.Draw(win, matrices[i])
		// }

		// draw the grid
		for x := 0; x < gridWidth; x++ {
			for y := 0; y < gridHeight; y++ {
				tile := tileGrid[x][y]
				if tile.Type == TYPE_WALL {
					imd.Color = pixel.RGB(0.153, 0.160, 0.196)
				} else {
					imd.Color = pixel.RGB(0.407, 0.54, 0.345)
				}
				// imd.Color = pixel.RGB(0, 0, 0)
				// draw bottom left of tile
				imd.Push(pixel.V(float64(x*tileSize), float64(y*tileSize)))

				// draw top right of tile
				imd.Push(pixel.V(float64((x+1)*tileSize), float64((y+1)*tileSize)))
				imd.Rectangle(0)
			}
		}

		imd.Color = pixel.RGB(0.8, 0.2, 0.2)
		imd.Push(pixel.V(float64(player.X), float64(player.Y)))
		imd.Push(pixel.V(float64(player.X+tileSize), float64(player.Y+(2*tileSize))))
		imd.Rectangle(0)

		imd.Draw(canvas)
		sprite := pixel.NewSprite(canvas, canvas.Bounds())
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()

		if win.JustPressed(pixel.KeySpace) {
			// iterate the grid
			tileGrid = iterateGrid(tileGrid)
		}
		if win.JustPressed(pixel.KeyR) {
			tileGrid = initGrid(tileGrid)
		}
	}
}

func main() {
	opengl.Run(run)
}
