package game

import "fmt"
import "math"
import rl "github.com/gen2brain/raylib-go/raylib"

// @fixme: scene is not being used because update requires the receiver to be
// pointer which gives a error when trying to use this interface
type Scene interface {
	Update(delta float32)
	Render()
}

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenTutorial
	ScreenGame
)

const (
	FontSizeS float32 = 24.0
	FontSizeM = 32.0
	FontSizeB = 40.0
	FontSizeL = 64.0
)

const gameName = "CHEXS"

// @note: not sure about this note
// @note: refers to the distance between tile center and borders, multiplies to
// two to get the distance adjacent tile point distance
const tileDefaultDistance float32 = 40.0
// @hack: I do not trust this to handle different environments
const tileDiagonalDistance float32 = tileDefaultDistance*2.0*1.72

const screenWidth = 720
const screenHeight = 720

const halfWidth = screenWidth/2
const halfHeight = screenHeight/2

var debug = false
var shouldQuit = false

var gameFont rl.Font
var pawnTexture rl.Texture2D

var target rl.RenderTexture2D
var cursor = rl.MouseCursorDefault

var foreground = rl.GetColor(0xCCCCCCFF)
var background = rl.GetColor(0x282828FF)

var gameScreen Game
var menuScreen = createMenu()
var tutorialScreen = createTutorial()

var currentScreen = ScreenMenu

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func play() {
	goToTutorial := true
	if goToTutorial {
		currentScreen = ScreenTutorial
	} else {
		startGame()
	}
}

func quit() {
	shouldQuit = true
}

func startGame() {
	gameScreen = createLevel1()
	currentScreen = ScreenGame
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
	default:        return ""
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

func measureText(text string, textSize float32) rl.Vector2 {
	return rl.MeasureTextEx(gameFont, text, textSize, 4.0)
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
	tile := &Tile{
		piece: piece,
		board: board,
		color: color,
		position: rl.Vector2{x, y},
	}
	if board == nil {
		return tile
	}
	board.tiles = append(board.tiles, tile)
	return board.tiles[len(board.tiles) - 1]
}

// @note: allow callback parameter allow us to include or not the passed tile
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

// @note: offseted position refers to tile position when being dragged around
func (tile Tile) offsetedPosition() rl.Vector2 {
	return rl.Vector2{
		tile.position.X + tile.moveOffset.X,
		tile.position.Y + tile.moveOffset.Y,
	}
}

func (tile Tile) isMoving() bool {
	return tile.moveOffset != rl.Vector2Zero()
}

func (tile Tile) vertices() []rl.Vector2 {
	verts := []rl.Vector2{}

	distance := tileDefaultDistance
	position := tile.offsetedPosition()

	x := position.X
	y := position.Y

	for i := 0; i < 6; i++ {
		angle := hexVertAngle(Direction(i))
		offsetX := float32(math.Cos(angle))*distance
		offsetY := float32(math.Sin(angle))*distance
		verts = append(verts, rl.Vector2{x + offsetX, y + offsetY})
	}
	return verts
}

func (tile Tile) onlyRender(highlight bool) {
	color := tile.color
	if highlight {
		color = rl.ColorBrightness(color, 0.2)
	}

	vertices := tile.vertices()
	for i := 0; i < len(vertices); i++ {
		pointA := vertices[i]

		pointBIndex := i + 1
		if pointBIndex >= len(vertices) {
			pointBIndex = 0
		}

		pointB := vertices[pointBIndex]
		rl.DrawLineEx(pointA, pointB, 2.0, color)
	}
}

func (tile Tile) onlyRenderPiece() {
	position := tile.offsetedPosition()

	position.X -= 25
	position.Y -= 28

	switch tile.piece {
	case Pawn: {
		rl.DrawTextureEx(pawnTexture, position, 0.0, 2.0, rl.White)
	}}
}

func (tile *Tile) Render(game *Game) {
	position := tile.offsetedPosition()
	highlight := tile == game.hoveredTile

	tile.onlyRender(highlight)

	// @note: debug center of tiles
	if debug {
		centerColor := rl.White
		if highlight {
			centerColor = rl.Red
		}
		rl.DrawCircleV(position, 2.0, centerColor)
	}

	color := tile.color
	if highlight {
		color = rl.ColorBrightness(color, 0.2)
	}

	tile.onlyRenderPiece()

	renderPossibleMoves := (
		tile.piece != Empty &&
		(tile == game.selectedTile || (game.selectedTile == nil && tile == game.hoveredTile)) &&
		game.mode == ModeSolve)

	if renderPossibleMoves {
		tiles := possibleMoves(tile)
		for _, it := range tiles {
			assert(it != tile, "Cannot highlight same tile")
			game.possibleMovesPoints = append(game.possibleMovesPoints, it.position)
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
		it.moveOffset = rl.Vector2Zero()
	}, true)
}

func (tile *Tile) cancelMove() {
	tile.iterateConnectedTiles(func (it *Tile) {
		it.moveOffset = rl.Vector2Zero()
	}, true)
}

func (tile *Tile) createNeighbor(direction Direction, piece Piece) *Tile {
	angle := hexFaceAngle(direction)

	neighborX := tile.position.X + float32(math.Cos(angle))*tileDefaultDistance*2.0
	neighborY := tile.position.Y + float32(math.Sin(angle))*tileDefaultDistance*2.0

	neighbor := createTile(tile.board, neighborX, neighborY, tile.color, piece)
	tile.neighbors[int(direction)] = neighbor

	opposite := direction.opposite()
	neighbor.neighbors[int(opposite)] = tile

	return neighbor
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

	titleSize := measureText(gameName, FontSizeL)
	titleX := halfWidth - titleSize.X/2
	titleY := halfHeight - titleSize.Y/2 - 64
	DrawText(gameName, titleX, titleY, FontSizeL, rl.White)

	author := "Johnathan"
	size := measureText(author, FontSizeS)
	x := screenWidth  - size.X     - 20.0
	y := screenHeight - size.Y*3.0 - 20.0
	DrawText(author, x, y, FontSizeS, rl.White)

	version := "raylib 6.x gamejam"
	size = measureText(version, FontSizeS)
	x = screenWidth  - size.X     - 20.0
	y = screenHeight - size.Y*2.0 - 20.0
	DrawText(version, x, y, FontSizeS, rl.White)

	pawnMention := "pawn by Master484"
	size = measureText(pawnMention, FontSizeS)
	x = screenWidth  - size.X     - 20.0
	y = screenHeight - size.Y*1.0 - 20.0
	DrawText(pawnMention, x, y, FontSizeS, rl.White)

	s.playButton.Render()
	s.quitButton.Render()
}

type TutorialPart int

const (
	TutorialIslands TutorialPart = iota
	TutorialSolving
	TutorialPawnMove
	TutorialPawnCapture
)

type Tutorial struct {
	part TutorialPart

	blink bool
	blinkTime float32
}

func createTutorial() Tutorial {
	return Tutorial{}
}

func (tutorial *Tutorial) Update(delta float32) {
	tutorial.blinkTime += delta
	if tutorial.blinkTime >= 0.5 {
		tutorial.blink = !tutorial.blink
		tutorial.blinkTime = 0.0
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		switch tutorial.part {
		case TutorialIslands:
			tutorial.part = TutorialSolving
		case TutorialSolving:
			tutorial.part = TutorialPawnMove
		case TutorialPawnMove:
			tutorial.part = TutorialPawnCapture
		case TutorialPawnCapture:
			tutorial.part = TutorialIslands
			startGame()
		}
	}
}

func renderPressSpace(blink bool) {
	text := "press space"
	size := measureText(text, FontSizeS)

	color := rl.White
	if blink {
		color = rl.Gray
	}

	x := screenWidth  - size.X - 20.0
	y := screenHeight - size.Y - 20.0

	DrawText(text, x, y, FontSizeS, color)
}

func (tutorial Tutorial) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()
	rl.ClearBackground(background)

	text := ""
	size := rl.Vector2{}

	switch tutorial.part {
	case TutorialIslands:
		renderPressSpace(tutorial.blink)

		text = "move the tiles"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight - size.Y*1.8, FontSizeM, rl.White)

		text = "build the puzzle"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight - size.Y*0.5, FontSizeM, rl.White)

		text = "and then solve it"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight + size.Y*0.8, FontSizeM, rl.White)
	case TutorialSolving:
		renderPressSpace(tutorial.blink)

		text = "to solve it you"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight - size.Y*1.8, FontSizeM, rl.White)

		text = "have to capture until"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight - size.Y*0.5, FontSizeM, rl.White)

		text = "a single piece remains"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, halfHeight + size.Y*0.8, FontSizeM, rl.White)
	case TutorialPawnMove:
		renderPressSpace(tutorial.blink)

		text = "pawn moves in faces"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, screenHeight*0.25, FontSizeM, rl.White)

		tile := createTile(nil, halfWidth, halfHeight*1.1, rl.White, Pawn)
		for i := 0; i < 6; i++ {
			direction := Direction(i)
			neighbor := tile.createNeighbor(direction, Empty)
			neighbor.color = rl.Green
		}

		tile.iterateConnectedTiles(func (tile *Tile) {
			tile.onlyRender(false)
			tile.onlyRenderPiece()
		}, true)
	case TutorialPawnCapture:
		renderPressSpace(tutorial.blink)

		text = "pawn captures in diagonals"
		size = measureText(text, FontSizeM)
		DrawText(text, halfWidth - size.X/2, screenHeight*0.25, FontSizeM, rl.White)

		tile := createTile(nil, halfWidth - tileDefaultDistance, halfHeight*1.25, rl.White, Pawn)
		tile.createNeighbor(Up, Empty)

		neighbor := tile.createNeighbor(UpRight, Empty)
		neighbor = neighbor.createNeighbor(Up, Empty)
		neighbor.color = rl.Green

		tile.iterateConnectedTiles(func (tile *Tile) {
			tile.onlyRender(false)
			tile.onlyRenderPiece()
		}, true)
	}
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

func (board *Board) isTileNeighbor(tileA, tileB *Tile) bool {
	tileAPosition := tileA.offsetedPosition()
	tileBPosition := tileB.offsetedPosition()

	distance := rl.Vector2Distance(tileAPosition, tileBPosition)
	return distance <= tileDefaultDistance*2.1 // @hack: 2.1 for error correction
}

func (board *Board) isTileDiagonal(tileA, tileB *Tile) bool {
	tileAPosition := tileA.offsetedPosition()
	tileBPosition := tileB.offsetedPosition()

	distanceMin := tileDiagonalDistance*0.9
	distanceMax := tileDiagonalDistance*1.1

	distance := rl.Vector2Distance(tileAPosition, tileBPosition)
	return distance >= distanceMin && distance <= distanceMax
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
	default:        return ""
	}
}

type State int

const (
	StateSolving State = iota
	StateFailed
	StateCompleted
)

func (state State) String() string {
	switch state {
	case StateSolving:   return "StateSolving"
	case StateFailed:    return "StateFailed"
	case StateCompleted: return "StateCompleted"
	default:			 return ""
	}
}

type Game struct {
	mode Mode
	state State
	board Board
	grid HexGrid
	movesLeft int
	// hovered tile used to highlight tile under cursor
	hoveredTile *Tile
	// selected tile refers to the first tile being moved
	selectedTile *Tile
	movingOrigin rl.Vector2
	// store possible move points to be draw after everything else
	possibleMovesPoints []rl.Vector2
}

func createGame() Game {
	grid := createGrid(screenWidth*0.25, screenHeight*0.25, 4, 5)
	grid.position = rl.Vector2{
		halfWidth  - grid.area.Width /2,
		halfHeight - grid.area.Height/2,
	}

	return Game{ grid: grid }
}

func createLevel1() Game {
	game := createGame()
	game.movesLeft = 1

	tilesA := createTile(&game.board, halfWidth*0.40, halfHeight, rl.Red, Pawn)
	for i := 0; i < 3; i++ {
		tilesA.createNeighbor(Direction(i), Empty)
	}

	tilesB := createTile(&game.board, halfWidth*1.60, halfHeight*0.40, rl.Blue, Pawn)
	for i := 0; i < 3; i++ {
		tilesB.createNeighbor(Direction(i*2), Empty)
	}

	return game
}

func (game Game) pieceCount() int {
	pieceCount := 0
	for _, tile := range game.board.tiles {
		if tile.piece != Empty {
			pieceCount += 1
		}
	}
	return pieceCount
}

func (game Game) isLevelFailed() bool {
	pieceCount := game.pieceCount()
	return game.movesLeft == 0 && pieceCount != 1
}

func (game Game) isLevelCompleted() bool {
	pieceCount := game.pieceCount()
	return game.movesLeft >= 0 && pieceCount == 1
}

func (game Game) isAnyTileColliding(tile *Tile) bool {
	tiles := map[*Tile]bool{}

	tile.iterateConnectedTiles(func (it *Tile) {
		tiles[it] = true
	}, true)

	collided := false
	tile.iterateConnectedTiles(func (it *Tile) {
		if collided {
			return
		}
		for _, itBoard := range game.board.tiles {
			_, ok := tiles[itBoard]
			if ok {
				continue
			}
			position := it.offsetedPosition()
			distance := rl.Vector2Distance(itBoard.position, position)
			if distance <= tileDefaultDistance*1.1 {
				collided = true
			}
		}
	}, true)

	return collided
}

func moveTilePiece(tile *Tile, newTile *Tile) bool {
	assert(tile != nil, "Tile is nil")
	assert(newTile != nil, "End tile is nil")

	moveMade := false
	for _, it := range possibleMoves(tile) {
		if it == newTile {
			moveMade = true
			newTile.piece = tile.piece
			tile.piece = Empty
		}
	}
	return moveMade
}

func possibleMoves(tile *Tile) []*Tile {
	tiles := []*Tile{}
	switch tile.piece {
	case Empty:
		return tiles
	case Pawn:
		board := tile.board
		for _, it := range board.tiles {
			if it == tile {
				continue
			}
			if board.isTileNeighbor(tile, it) {
				if it.piece == Empty {
					tiles = append(tiles, it)
				}
			}
			if board.isTileDiagonal(tile, it) {
				if it.piece != Empty {
					tiles = append(tiles, it)
				}
			}
		}
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

func (game *Game) Update(_delta float32) {
	if game.state == StateFailed {
		if rl.IsKeyPressed(rl.KeySpace) {
			gameScreen = createLevel1()
		}
		return
	}

	if game.state == StateCompleted {
		if rl.IsKeyPressed(rl.KeySpace) {
			currentScreen = ScreenMenu
		}
		return
	}

	mousePosition := rl.GetMousePosition()
	game.hoveredTile = selectTile(game.board.tiles, mousePosition)

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		switch game.mode {
		case ModeBuild:
			if game.selectedTile == nil {
				game.movingOrigin = rl.GetMousePosition()
			}
			game.selectedTile = game.hoveredTile
			if game.selectedTile == nil {
				game.movingOrigin = rl.Vector2Zero()
			}
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		switch game.mode {
		case ModeBuild:
			if game.selectedTile != nil {
				tile := game.selectedTile

				closestSnap := closestSnapPoint(tile, game.grid)
				snapOffset := rl.Vector2Subtract(closestSnap, tile.position)

				tile.move(snapOffset.X, snapOffset.Y)
				isTileCollinding := game.isAnyTileColliding(tile)
				if isTileCollinding {
					tile.cancelMove()
				} else {
					tile.applyMove()
				}
			}
			game.movingOrigin = rl.Vector2Zero()
			game.selectedTile = nil
		case ModeSolve:
			if game.selectedTile == nil {
				if game.hoveredTile != nil && game.hoveredTile.piece != Empty {
					game.selectedTile = game.hoveredTile
				}
			} else if game.hoveredTile == nil {
				game.selectedTile = nil
			} else {
				if game.selectedTile != game.hoveredTile {
					if moveTilePiece(game.selectedTile, game.hoveredTile) {
						game.selectedTile = game.hoveredTile
						game.hoveredTile = nil
						game.movesLeft -= 1
						if game.isLevelFailed() {
							game.state = StateFailed
						} else if game.isLevelCompleted() {
							game.state = StateCompleted
						}
					}
				}
			}
		}
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		switch game.mode {
		case ModeBuild:
			game.mode = ModeSolve
		case ModeSolve:
			game.mode = ModeBuild
			if game.selectedTile != nil {
				game.selectedTile.cancelMove()
				game.selectedTile = nil
			}
		}
	}

	if game.selectedTile != nil {
		switch game.mode {
		case ModeBuild:
			moveDistance := rl.Vector2Subtract(mousePosition, game.movingOrigin)
			if game.selectedTile != nil {
				game.selectedTile.move(moveDistance.X, moveDistance.Y)
			}
		}
	}
}

func (game Game) renderLevelFailed() {
	screenArea := rl.Rectangle{0, 0, screenWidth, screenHeight}
	overlayColor := rl.ColorAlpha(rl.Black, 0.8)
	rl.DrawRectangleRec(screenArea, overlayColor)

	text := "you failed!"
	size := measureText(text, FontSizeB)
	DrawText(text, halfWidth - size.X/2, halfHeight - size.Y/2, FontSizeB, foreground)

	text = "space to retry"
	size = measureText(text, FontSizeS)
	DrawText(text, halfWidth - size.X/2, halfHeight + size.Y*1.1, FontSizeS, foreground)
}

func (game Game) renderLevelCompleted() {
	screenArea := rl.Rectangle{0, 0, screenWidth, screenHeight}
	overlayColor := rl.ColorAlpha(rl.Black, 0.8)
	rl.DrawRectangleRec(screenArea, overlayColor)

	text := "you completed!"
	size := measureText(text, FontSizeB)
	DrawText(text, halfWidth - size.X/2, halfHeight - size.Y/2, FontSizeB, foreground)

	text = "space to continue"
	size = measureText(text, FontSizeS)
	DrawText(text, halfWidth - size.X/2, halfHeight + size.Y*1.1, FontSizeS, foreground)
}

func (game Game) Render() {
	rl.BeginTextureMode(target)
	defer rl.EndTextureMode()
	rl.ClearBackground(background)

	if debug {
		LiveInfoFrameReset()

		infoFps := fmt.Sprint("fps: ", rl.GetFPS())
		LiveInfo(infoFps)

		infoTiles := fmt.Sprint("tiles: ", len(game.board.tiles))
		LiveInfo(infoTiles)

		infoArea := fmt.Sprint("area: ", game.grid.area)
		LiveInfo(infoArea)

		infoMode := fmt.Sprint("mode: ", game.mode)
		LiveInfo(infoMode)

		infoHovered := fmt.Sprint("hovered: ", game.hoveredTile)
		LiveInfo(infoHovered)

		infoSelected := fmt.Sprint("selected: ", game.selectedTile)
		LiveInfo(infoSelected)

		infoSnap := "snap: No tile selected"
		LiveInfo(infoSnap)

		area := game.grid.area
		area.X += game.grid.position.X
		area.Y += game.grid.position.Y
		rl.DrawRectangleLinesEx(area, 2.0, rl.Red)
	}

	game.possibleMovesPoints = nil

	game.grid.Render()
	for _, tile := range game.board.tiles {
		tile.Render(&game)
	}

	for _, point := range game.possibleMovesPoints {
		rl.DrawCircleV(point, 4.0, rl.Green)
	}

	size := rl.Vector2{}
	modeText := ""

	switch game.mode {
	case ModeBuild:
		modeText   = "build mode"
	case ModeSolve:
		modeText   = "solve mode"
	}

	size = measureText(modeText, FontSizeM)
	DrawText(modeText, halfWidth - size.X/2, screenHeight*0.8, FontSizeM, rl.White)

	text := "space to switch"
	size = measureText(text, FontSizeS)
	DrawText(text, halfWidth - size.X/2, screenHeight*0.8 + FontSizeM*1.1, FontSizeS, rl.White)

	movesLeft := fmt.Sprint("moves left: ", game.movesLeft)
	DrawText(movesLeft, 20.0, 20.0, FontSizeM, rl.White)

	if game.state == StateFailed {
		game.renderLevelFailed()
		return
	}

	if game.state == StateCompleted {
		game.renderLevelCompleted()
		return
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
	case ScreenMenu:
		menuScreen.Update(delta)
		menuScreen.Render()
	case ScreenGame:
		gameScreen.Update(delta)
		gameScreen.Render()
	case ScreenTutorial:
		tutorialScreen.Update(delta)
		tutorialScreen.Render()
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

	gameFont = rl.LoadFontEx("./assets/arvo/Arvo-Bold.ttf", 96, nil, 250)
	defer rl.UnloadFont(gameFont)

	pawnTexture = rl.LoadTexture("./assets/pieces/pawn.png")
	defer rl.UnloadTexture(pawnTexture)

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
