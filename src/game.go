package game

import "fmt"
import "math"
import rl "github.com/gen2brain/raylib-go/raylib"

type Screen interface {
	Update(delta float32)
	Render()
}

type ScreenIndex int

const (
	ScreenIndexMenu ScreenIndex = iota
	ScreenIndexGame
)

const gameName = "CHEXS"
const tileDefaultDistance float32 = 40.0

const screenWidth = 720
const screenHeight = 720

const halfWidth = screenWidth/2
const halfHeight = screenHeight/2

var gameFont rl.Font
var shouldQuit = false

var target rl.RenderTexture2D
var cursor = rl.MouseCursorDefault

var background = rl.GetColor(0x282828FF)

var menuScreen = createMenu()
var gameScreen = createGame()

var currentScreen = ScreenIndexGame

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

type Direction int

const (
	Down Direction = iota
	DownLeft
	UpLeft
	Up
	UpRight
	DownRight
)

func (direction Direction) String() string {
	switch direction {
	case Down:      return "Down"
	case DownLeft:  return "DownLeft"
	case UpLeft:    return "UpLeft"
	case Up:        return "Up"
	case UpRight:   return "UpRight"
	case DownRight: return "DownRight"
	default: return ""
	}
}

func (direction Direction) opposite() Direction {
	return Direction((int(direction) + 3) % (int(DownRight) + 1))
}

func hexVertAngle(d Direction) float64 {
	return math.Pi*2.0/6.0*float64(d)
}

func hexFaceAngle(d Direction) float64 {
	return math.Pi*2.0/6.0*float64(d) + math.Pi*0.5
}

func drawHexTile(x, y, distance float32, color rl.Color) {
	verts := []rl.Vector2{}
	faces := []rl.Vector2{}

	for i := 0; i < 6; i++ {
		angle := hexVertAngle(Direction(i))
		offsetX := float32(math.Cos(angle))*distance
		offsetY := float32(math.Sin(angle))*distance
		verts = append(verts, rl.Vector2{x + offsetX, y + offsetY})

		tAngle := hexFaceAngle(Direction(i))
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
}

type Tile struct {
	piece int
	board *Board
	color rl.Color
	position rl.Vector2
	moveOffset rl.Vector2
	moveOffsetApplied bool
	neighbors [6]*Tile
	visited bool // whether tile was visited in a iteration
}

func createTile(board *Board, x, y float32, color rl.Color, piece int) *Tile {
	board.tiles = append(board.tiles, &Tile{
		board: board,
		color: color,
		position: rl.Vector2{x, y},
	})
	return board.tiles[len(board.tiles) - 1]
}

func (tile *Tile) iterateConnectedTiles(callback func (*Tile)) {
	if tile.visited {
		return
	}
	callback(tile)
	tile.visited = true
	for _, neighbor := range tile.neighbors {
		if neighbor == nil {
			continue
		}
		neighbor.iterateConnectedTiles(callback)
	}
	tile.visited = false
}

func (tile Tile) offsetedPosition() rl.Vector2 {
	return rl.Vector2{
		tile.position.X + tile.moveOffset.X,
		tile.position.Y + tile.moveOffset.Y,
	}
}

func (tile Tile) DrawTile(highlight bool) {
	color := tile.color
	if highlight {
		color = rl.ColorBrightness(color, 0.2)
	}
	position := tile.offsetedPosition()
	drawHexTile(position.X, position.Y, tileDefaultDistance, color)
}

func (tile *Tile) move(x, y float32) {
	tile.iterateConnectedTiles(func (tile *Tile) {
		tile.moveOffset.X = x
		tile.moveOffset.Y = y
	})
}

func (tile *Tile) applyMove() {
	tile.iterateConnectedTiles(func (tile *Tile) {
		tile.position.X += tile.moveOffset.X
		tile.position.Y += tile.moveOffset.Y
		tile.cancelMove()
	})
}

func (tile *Tile) cancelMove() {
	tile.moveOffset = rl.Vector2Zero()
}

func (tile *Tile) createNeighbor(direction Direction, piece int) {
	angle := hexFaceAngle(direction)

	neighborX := tile.position.X + float32(math.Cos(angle))*tileDefaultDistance*2.0
	neighborY := tile.position.Y + float32(math.Sin(angle))*tileDefaultDistance*2.0

	neighbor := createTile(tile.board, neighborX, neighborY, tile.color, piece)
	tile.neighbors[int(direction)] = neighbor

	opposite := direction.opposite()
	neighbor.neighbors[int(opposite)] = tile
}

type Board struct {
	tiles []*Tile
}

type Menu struct {
	playButton Button
	quitButton Button
}

func createMenu() Menu {
	var x, y float32 

	x = halfWidth - 180/2
	y = halfHeight - 64/2 + 32

	playButton := CreateButton("PLAY", x, y, 180, 48); y += 64
	playButton.Callback = play

	quitButton := CreateButton("QUIT", x, y, 180, 48)
	quitButton.Callback = quit

	return Menu{ playButton, quitButton }
}

func (s Menu) Update(_delta float32) {}

func (s Menu) Render() {
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

type HexGrid struct {
	cols int32
	rows int32
	area rl.Rectangle
	// the current way of generating hexagon grids defines points outside of
	// the area they should be, which make them negative, this offset is used
	// to render them in the right area
	areaOffset float32
	// @todo: we could replace position with area 'x and 'y
	position rl.Vector2
	points []rl.Vector2
}

func createGrid(x, y float32, cols, rows int32) HexGrid {
	point := rl.Vector2{}
	points := []rl.Vector2{}
	directionIndex := 0
	directions := []Direction{ UpRight, DownRight }
	for i := int32(0); i < cols; i++ {
		for j := int32(0); j < rows; j++ {
			points = append(points, point)
			angle := hexFaceAngle(directions[directionIndex])
			point.X += float32(math.Cos(angle))*tileDefaultDistance*2.0
			point.Y += float32(math.Sin(angle))*tileDefaultDistance*2.0
			directionIndex = (directionIndex + 1) % 2
		}
		directionIndex = 0
		point = rl.Vector2Zero()
		angle := hexFaceAngle(Down)
		point.X += float32(math.Cos(angle))*tileDefaultDistance*float32(i + 1)*2.0
		point.Y += float32(math.Sin(angle))*tileDefaultDistance*float32(i + 1)*2.0
	}
	area, offset := gridAreaFromPoints(points)
	return HexGrid{
		cols: cols,
		rows: rows,
		area: area,
		areaOffset: offset,
		points: points,
		position: rl.Vector2{ x, y },
	}
}

// @todo: better define this function, it's not clear why it should return
// two values
func gridAreaFromPoints(points []rl.Vector2) (rl.Rectangle, float32) {
	var minX, minY, maxX, maxY float32
	for _, point := range points {
		if point.X < minX {
			minX = point.X
		} else if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		} else if point.Y > maxY {
			maxY = point.Y
		}
	}
	offset := float32(math.Abs(float64(minY)))
	assert(minY < 0, "Minimum Y should be negative")
	minY = 0
	maxY += offset 
	return rl.Rectangle{
		minX, minY,
		minX + maxX,
		minY + maxY,
	}, offset
}

func (grid HexGrid) Render() {
	for _, point := range grid.points {
		rl.DrawCircle(
			int32(point.X) + int32(grid.position.X),
			int32(point.Y) + int32(grid.position.Y) + int32(grid.areaOffset),
			2.0, rl.White)
	}
}

type Game struct {
	board Board
	grid HexGrid
	hoveredTile *Tile  // hovered tile used to highlight tile under cursor
	selectedTile *Tile // selected tile refers to the first tile being moved
	movingOrigin rl.Vector2
}

func createGame() Game {
	board := Board{}

	tilesA := createTile(&board, halfWidth*1.5, halfHeight*1.5, rl.Red, 1)
	for i := 0; i < 3; i++ {
		tilesA.createNeighbor(Direction(i), 1)
	}

	tilesB := createTile(&board, halfWidth*0.5, halfHeight*0.5, rl.Blue, 2)
	for i := 0; i < 3; i++ {
		tilesB.createNeighbor(Direction(i*2), 1)
	}

	grid := createGrid(screenWidth*0.25, screenHeight*0.25, 5, 5)
	grid.position = rl.Vector2{
		halfWidth - grid.area.Width/2,
		halfHeight - grid.area.Height/2,
	}

	return Game{
		grid: grid,
		board: board,
	}
}

func selectTile(tiles []*Tile, mousePosition rl.Vector2) *Tile {
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		tilePosition := tile.offsetedPosition() // offset position needed when dragging
		if rl.Vector2Distance(mousePosition, tilePosition) < tileDefaultDistance {
			return tile
		}
	}
	return nil
}

func (s *Game) Update(_delta float32) {
	mousePosition := rl.GetMousePosition()
	s.hoveredTile = selectTile(s.board.tiles, mousePosition)

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		if s.selectedTile == nil {
			s.movingOrigin = rl.GetMousePosition()
		}
		s.selectedTile = selectTile(s.board.tiles, mousePosition)
		if s.selectedTile == nil {
			s.movingOrigin = rl.Vector2Zero()
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		if s.selectedTile != nil {
			s.selectedTile.applyMove()
		}
		s.movingOrigin = rl.Vector2Zero()
		s.selectedTile = nil
	}

	if s.selectedTile != nil {
		moveDistance := rl.Vector2Subtract(mousePosition, s.movingOrigin)
		if s.selectedTile != nil {
			s.selectedTile.move(moveDistance.X, moveDistance.Y)
		}
	}
}

func (s Game) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()
	rl.ClearBackground(background)

	LiveInfoFrameReset()
	LiveInfo(fmt.Sprint("fps: ", rl.GetFPS()))
	LiveInfo(fmt.Sprint("tiles: ", len(s.board.tiles)))
	LiveInfo(fmt.Sprint("grid area: ", s.grid.area))

	area := s.grid.area
	area.X += s.grid.position.X
	area.Y += s.grid.position.Y
	rl.DrawRectangleLinesEx(area, 2.0, rl.Red)

	s.grid.Render()
	for _, tile := range s.board.tiles {
		highlight := tile == s.hoveredTile
		tile.DrawTile(highlight)
	}
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

func updateCurrentScreen(delta float32) {
	switch currentScreen {
	case ScreenIndexMenu:
		menuScreen.Update(delta)
		menuScreen.Render()
	case ScreenIndexGame:
		gameScreen.Update(delta)
		gameScreen.Render()
	default:
		assert(false, "Unreachable")
	}
}

func updateFrame() {
	delta := rl.GetFrameTime()
	cursor = rl.MouseCursorDefault
	updateCurrentScreen(delta)
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
