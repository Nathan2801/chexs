package game

import _ "fmt"
import "math"
import rl "github.com/gen2brain/raylib-go/raylib"

type Renderable interface {
	Render()
}

type ScreenIndex int

const (
	ScreenIndexMenu ScreenIndex = iota
	ScreenIndexGame
)

const gameName = "CHEXS"

const screenWidth = 720
const screenHeight = 720

const halfWidth = screenWidth/2
const halfHeight = screenHeight/2

var gameFont rl.Font
var shouldQuit = false
var target rl.RenderTexture2D
var currentScreen = ScreenIndexMenu
var background = rl.GetColor(0x282828FF)
var cursor = rl.MouseCursorDefault
var screens = []Renderable{
	createScreenMenu(),
	createScreenGame(),
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func play() {
	currentScreen = ScreenIndexGame
}

func quit() {
	shouldQuit = true
}

type ScreenMenu struct {
	playButton Button
	quitButton Button
}

func createScreenMenu() ScreenMenu {
	var x, y float32 

	x = halfWidth - 180/2
	y = halfHeight - 64/2 + 32

	playButton := CreateButton("PLAY", x, y, 180, 48); y += 64
	playButton.Callback = play

	quitButton := CreateButton("QUIT", x, y, 180, 48)
	quitButton.Callback = quit

	return ScreenMenu{ playButton, quitButton }
}

func (s ScreenMenu) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()

	rl.ClearBackground(background)

	// Draw title
	titleFontSize := float32(64.0)
	titleSize := rl.MeasureTextEx(gameFont, gameName, titleFontSize, 4)

	titleX := halfWidth - titleSize.X/2
	titleY := halfHeight - titleSize.Y/2 - 64

	DrawText(gameName, titleX, titleY, int(titleFontSize), rl.White)

	// Draw buttons
	s.playButton.Render()
	s.quitButton.Render()
}

type ScreenGame struct {}

func createScreenGame() ScreenGame {
	return ScreenGame{}
}

func hexVertAngle(x, y float32, i int) float64 {
	return math.Pi*2.0/6.0*float64(i)
}

func hexFaceAngle(x, y float32, i int) float64 {
	return math.Pi*2.0/6.0*float64(i) + math.Pi*0.5
}

func drawHexTile(x, y, distance float32, color rl.Color, neighbors bool) {
	verts := []rl.Vector2{}
	faces := []rl.Vector2{}

	for i := 0; i < 6; i++ {
		angle := hexVertAngle(x, y, i)
		offsetX := float32(math.Cos(angle))*distance
		offsetY := float32(math.Sin(angle))*distance
		verts = append(verts, rl.Vector2{x + offsetX, y + offsetY})

		tAngle := hexFaceAngle(x, y, i)
		tX := float32(math.Cos(tAngle))*distance*2.0
		tY := float32(math.Sin(tAngle))*distance*2.0
		faces = append(faces, rl.Vector2{x + tX, y + tY})
	}

	rl.DrawCircle(int32(x), int32(y), 2.0, rl.White)
	for i := 0; i < len(verts); i++ {
		pointA := verts[i]

		pointBIndex := i + 1
		if pointBIndex >= len(verts) {
			pointBIndex = 0
		}

		pointB := verts[pointBIndex]
		rl.DrawLineEx(pointA, pointB, 2.0, color)
	}

	if neighbors {
		for i := 0; i < len(faces); i++ {
			face := faces[i]
			drawHexTile(face.X, face.Y, distance, color, false)
		}
	}
}

func (s ScreenGame) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()

	drawHexTile(float32(halfWidth), float32(halfHeight), 40.0, rl.Red, true)

	rl.ClearBackground(background)
}

func renderTargetTexture() {
	var r0 = rl.Rectangle {
		0, 0,
		float32(target.Texture.Width),
		-float32(target.Texture.Height),
	}

	var r1 = rl.Rectangle{
		0, 0,
		float32(target.Texture.Width),
		float32(target.Texture.Height),
	}

	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.White)
	rl.DrawTexturePro(target.Texture, r0, r1, rl.Vector2{}, 0.0, rl.White)
}

func updateFrame() {
	cursor = rl.MouseCursorDefault
	assert(int(currentScreen) >= 0 || int(currentScreen) < len(screens),
		"Invalid screen")
	// note: render method does much more than only rendering it's the update
	//       and rendering methods at the same time.
	screens[currentScreen].Render()
	rl.SetMouseCursor(cursor)
	renderTargetTexture()
}

func RunGame() {
	rl.InitWindow(screenWidth, screenHeight, "Game")
	defer rl.CloseWindow()

	gameFont = rl.LoadFontEx("./assets/arvo/Arvo-Bold.ttf", 64, nil, 250)
	defer rl.UnloadFont(gameFont)

	SetUIFont(&gameFont)

	target = rl.LoadRenderTexture(screenWidth, screenHeight)
	defer rl.UnloadRenderTexture(target)

	rl.SetTextureFilter(target.Texture, rl.FilterBilinear)
	//rl.SetMainLoop(updateFrame)

	for !rl.WindowShouldClose() {
		if shouldQuit {
			break
		}
		updateFrame()
	}
}
