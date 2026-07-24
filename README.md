
# Chexs for Raylib 6.x GameJam

A simple chess puzzle game written for the [Raylib 6.x GameJam](https://itch.io/jam/raylib-6x-gamejam).

:video_game: You can play Chexs [here](https://sokobo.itch.io/chexs).

## Theme

The Jam theme was _merge + hexagons_ which got me thinking about a chess-puzzle
game, but instead of squares they move in hexagons, then I thought how could I
introduce the _merge_ part, it leads me to simply jigsaw puzzles, which gave
me the idea of _"merging"_ islands of hexagons to build your own solution, and
then solving it under the constraint of limited moves.

## Instructions

Your goal is to build the puzzle and then with a limited amount of moves
captures all pieces until a single one remains.

Island are group of tiles that contains pieces, in build mode, click and drag
islands to build the puzzle.

After the puzzle is built you can press <kbd>Space</kbd> to switch to solve
mode, then you can move pieces, but after a single move is done you cannot go
into build mode anymore.

In this version there is only pawns pieces, pawns moves towards hexagons faces,
and captures in "diagonals", a hexagon diagonal is the closest tile pointed by
a hexagon vertice.

## Building Project

### For Windows Using Powershell

```powershell
git clone https://github.com/Nathan2801/chexs/
cd chexs
git submodule init   # Required for Raylib-Go-Wasm
git submodule update # Required for Raylib-Go-Wasm
.\make.ps1
```

### For Browser Using Powershell

```powershell
git clone https://github.com/Nathan2801/chexs/
cd chexs
git submodule init   # Required for Raylib-Go-Wasm
git submodule update # Required for Raylib-Go-Wasm
```

> [!NOTE]
> You should uncomment one line in `src/game.go` to enable raylib SetMainLoop
> function, and two lines in `main.go` to enable assets loading. You can search
> for: `Uncomment this for wasm`.

```powershell
.\make.ps1 wasm
```

## Thanks

- [Master484](https://opengameart.org/users/master484) for the [pawn sprite](https://opengameart.org/content/m484-chess-set).

