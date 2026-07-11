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

type Piece int

const (
	Empty Piece = iota
	Pawn
)

func hexVertAngle(d Direction) float64 {
	return math.Pi*2.0/6.0*float64(d)
}

func hexFaceAngle(d Direction) float64 {
	return math.Pi*2.0/6.0*float64(d) + math.Pi*0.5
}

type Tile struct {
	piece Piece
	board *Board
	color rl.Color
	position rl.Vector2
	moveOffset rl.Vector2
	moveOffsetApplied bool
	neighbors [6]*Tile
	visited bool // whether tile was visited in a iteration
}

func createTile(board *Board, x, y float32, color rl.Color, piece Piece) *Tile {
	board.tiles = append(board.tiles, &Tile{
		piece: piece,
		board: board,
		color: color,
		position: rl.Vector2{x, y},
	})
	return board.tiles[len(board.tiles) - 1]
}

// @note: allow callback parameter allow us to include or exclude the passed
// tile
func (tile *Tile) iterateConnectedTiles(callback func (*Tile), allowCallback bool) {
	if tile.visited {
		return
	}
	if allowCallback {
		callback(tile)
	}
	tile.visited = true
	for _, neighbor := range tile.neighbors {
		if neighbor == nil {
			continue
		}
		neighbor.iterateConnectedTiles(callback, true)
	}
	tile.visited = false
}

// offseted position refers to tile position when it is being dragged around
func (tile Tile) offsetedPosition() rl.Vector2 {
	return rl.Vector2{
		tile.position.X + tile.moveOffset.X,
		tile.position.Y + tile.moveOffset.Y,
	}
}

func (tile Tile) isMoving() bool {
	return tile.moveOffset != rl.Vector2Zero()
}

func (tile *Tile) Render(highlight bool, gameMode Mode) {
	verts := []rl.Vector2{}
	faces := []rl.Vector2{}

	distance := tileDefaultDistance
	position := tile.offsetedPosition()

	x := position.X
	y := position.Y

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

	centerColor := rl.White
	if highlight {
		centerColor = rl.Red
	}
	rl.DrawCircleV(position, 2.0, centerColor)

	color := tile.color
	if highlight {
		color = rl.ColorBrightness(color, 0.2)
	}

	for i := 0; i < len(verts); i++ {
		pointA := verts[i]

		pointBIndex := i + 1
		if pointBIndex >= len(verts) {
			pointBIndex = 0
		}

		pointB := verts[pointBIndex]
		rl.DrawLineEx(pointA, pointB, 2.0, color)
	}

	switch tile.piece {
	case Pawn:
		rl.DrawRectangleRec(rl.Rectangle{x - 10, y - 10, 20, 20}, rl.White)
	}

	isSolveMode := gameMode == ModeSolve
	mouseOverPiece := highlight && tile.piece != Empty
	if mouseOverPiece && isSolveMode && !tile.isMoving() {
		tiles := possibleMoves(tile)
		for _, it := range tiles {
			assert(it != tile, "Cannot highlight same tile")
			rl.DrawCircleV(it.position, 4.0, rl.Green)
		}
	}
}

func (tile *Tile) move(x, y float32) {
	tile.iterateConnectedTiles(func (it *Tile) {
		it.moveOffset.X = x
		it.moveOffset.Y = y
	}, true)
}

func (tile *Tile) applyMove() {
	tile.iterateConnectedTiles(func (it *Tile) {
		it.position.X += it.moveOffset.X
		it.position.Y += it.moveOffset.Y
		it.cancelMove()
	}, true)
}

func (tile *Tile) cancelMove() {
	tile.moveOffset = rl.Vector2Zero()
}

func (tile *Tile) createNeighbor(direction Direction, piece Piece) {
	angle := hexFaceAngle(direction)

	neighborX := tile.position.X + float32(math.Cos(angle))*tileDefaultDistance*2.0
	neighborY := tile.position.Y + float32(math.Sin(angle))*tileDefaultDistance*2.0

	neighbor := createTile(tile.board, neighborX, neighborY, tile.color, piece)
	tile.neighbors[int(direction)] = neighbor

	opposite := direction.opposite()
	neighbor.neighbors[int(opposite)] = tile
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

// @todo: fix naming convention, here we specify HexGrid but tile we call it
// Tile not HexTile
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

func (grid HexGrid) pointPosition(point rl.Vector2) rl.Vector2 {
	return rl.Vector2{
		point.X + grid.position.X,
		point.Y + grid.position.Y + grid.areaOffset,
	}
}

func (grid HexGrid) Render() {
	for _, point := range grid.points {
		pointPosition := grid.pointPosition(point)
		rl.DrawCircle(
			int32(pointPosition.X),
			int32(pointPosition.Y),
			2.0, rl.White)
	}
}

// @note: even tho board has a single element we define it like this so tile
// can have a pointer to a board instead of having a pointer to a list of
// pointer tiles
type Board struct {
	tiles []*Tile
}

type Mode int

const (
	ModeBuild Mode = iota
	ModeSolve
)

func (mode Mode) String() string {
	switch mode {
	case ModeBuild: return "ModeBuild"
	case ModeSolve: return "ModeSolve"
	default: return ""
	}
}

type Game struct {
	mode Mode
	board Board
	tiles []*Tile
	grid HexGrid
	hoveredTile *Tile  // hovered tile used to highlight tile under cursor
	selectedTile *Tile // selected tile refers to the first tile being moved
	movingOrigin rl.Vector2
}

func createGame() Game {
	board := Board{}

	tilesA := createTile(&board, halfWidth*1.5, halfHeight*1.5, rl.Red, Empty)
	for i := 0; i < 3; i++ {
		tilesA.createNeighbor(Direction(i), Empty)
	}

	tilesB := createTile(&board, halfWidth*0.5, halfHeight*0.5, rl.Blue, Pawn)
	for i := 0; i < 3; i++ {
		tilesB.createNeighbor(Direction(i*2), Empty)
	}

	grid := createGrid(screenWidth*0.25, screenHeight*0.25, 5, 5)
	grid.position = rl.Vector2{
		halfWidth - grid.area.Width/2,
		halfHeight - grid.area.Height/2,
	}

	return Game{
		mode: ModeBuild,
		grid: grid,
		board: board,
	}
}

func possibleMoves(tile *Tile) []*Tile {
	tiles := []*Tile{}
	switch tile.piece {
	case Empty:
		return tiles
	case Pawn:
		tile.iterateConnectedTiles(func (neighbor *Tile) {
			tiles = append(tiles, neighbor)
		}, false)
	}
	return tiles
}

func selectTile(tiles []*Tile, mousePosition rl.Vector2) *Tile {
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		tilePosition := tile.offsetedPosition()
		if rl.Vector2Distance(mousePosition, tilePosition) < tileDefaultDistance {
			return tile
		}
	}
	return nil
}

func closestSnapPoint(tile *Tile, grid HexGrid) rl.Vector2 {
	pointFound := false
	closestPoint := rl.Vector2Zero()
	closestDistance := float32(math.MaxFloat32)
	for _, point := range grid.points {
		tilePosition := tile.offsetedPosition()
		pointPosition := grid.pointPosition(point)
		distance := rl.Vector2Distance(tilePosition, pointPosition)
		if distance < closestDistance {
			closestPoint = pointPosition
			closestDistance = distance
		}
	}
	assert(!pointFound, "Point not found")
	return closestPoint
}

func (s *Game) Update(_delta float32) {
	mousePosition := rl.GetMousePosition()
	s.hoveredTile = selectTile(s.board.tiles, mousePosition)

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		switch s.mode {
		case ModeBuild:
			if s.selectedTile == nil {
				s.movingOrigin = rl.GetMousePosition()
			}
			s.selectedTile = selectTile(s.board.tiles, mousePosition)
			if s.selectedTile == nil {
				s.movingOrigin = rl.Vector2Zero()
			}
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		switch s.mode {
		case ModeBuild:
			tile := s.selectedTile
			if tile != nil {
				closestSnap := closestSnapPoint(tile, s.grid)
				snapOffset := rl.Vector2Subtract(closestSnap, tile.position)
				s.selectedTile.move(snapOffset.X, snapOffset.Y)
				s.selectedTile.applyMove()
			}
			s.movingOrigin = rl.Vector2Zero()
			s.selectedTile = nil
		}
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		switch s.mode {
		case ModeBuild:
			s.mode = ModeSolve
		case ModeSolve:
			s.mode = ModeBuild
		}
	}

	if s.selectedTile != nil {
		switch s.mode {
		case ModeBuild:
			moveDistance := rl.Vector2Subtract(mousePosition, s.movingOrigin)
			if s.selectedTile != nil {
				s.selectedTile.move(moveDistance.X, moveDistance.Y)
			}
		}
	}
}

func (s Game) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()
	rl.ClearBackground(background)

	LiveInfoFrameReset()

	infoFps := fmt.Sprint("fps: ", rl.GetFPS())
	LiveInfo(infoFps)

	infoTiles := fmt.Sprint("tiles: ", len(s.board.tiles))
	LiveInfo(infoTiles)

	infoArea := fmt.Sprint("area: ", s.grid.area)
	LiveInfo(infoArea)

	infoMode := fmt.Sprint("mode: ", s.mode)
	LiveInfo(infoMode)

	infoSnap := "snap: No tile selected"
	if s.selectedTile != nil {
		closestSnap := closestSnapPoint(s.selectedTile, s.grid)
		infoSnap = fmt.Sprint("snap: ", closestSnap)
		rl.DrawCircleV(closestSnap, 4.0, rl.Yellow)
	}
	LiveInfo(infoSnap)

	area := s.grid.area
	area.X += s.grid.position.X
	area.Y += s.grid.position.Y
	rl.DrawRectangleLinesEx(area, 2.0, rl.Red)

	s.grid.Render()
	for _, tile := range s.board.tiles {
		highlight := tile == s.hoveredTile
		tile.Render(highlight, s.mode)
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
