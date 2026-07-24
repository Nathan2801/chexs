module game

go 1.26.4

require github.com/gen2brain/raylib-go/raylib v0.60.0

require (
	github.com/BrownNPC/Raylib-Go-Wasm/wasm-runtime v0.0.0-00010101000000-000000000000 // indirect
	github.com/BrownNPC/wasm-ffi-go v1.3.0 // indirect
)

replace (
	github.com/BrownNPC/Raylib-Go-Wasm/wasm-runtime => ./external/Raylib-Go-Wasm/wasm-runtime
	github.com/gen2brain/raylib-go/raygui => ./external/Raylib-Go-Wasm/raygui
	github.com/gen2brain/raylib-go/raylib => ./external/Raylib-Go-Wasm/raylib
)
