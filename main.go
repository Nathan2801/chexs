package main

import "embed"
import "game/src"
// Uncomment this for wasm.
import rl "github.com/gen2brain/raylib-go/raylib"

// REQUIRED CODE FOR LOADING ASSETS ON WEB
//
//go:embed assets/*
var ASSETS embed.FS

func init() {
	// Uncomment this for wasm.
	rl.AddFileSystem(ASSETS)
}

func main() {
	game.RunGame()
}
