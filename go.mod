module game

go 1.26.4

require github.com/gen2brain/raylib-go/raylib v0.60.0

require (
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/jupiterrider/ffi v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
)

replace (
	github.com/BrownNPC/Raylib-Go-Wasm/wasm-runtime => ./external/Raylib-Go-Wasm/wasm-runtime
	github.com/gen2brain/raylib-go/raygui => ./external/Raylib-Go-Wasm/raygui
//github.com/gen2brain/raylib-go/raylib => ./external/Raylib-Go-Wasm/raylib
)
