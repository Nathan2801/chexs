package main

import "embed"
import "game/src"
import _ "github.com/gen2brain/raylib-go/raylib"

// REQUIRED CODE FOR LOADING ASSETS ON WEB
//
//go:embed assets/*
var ASSETS embed.FS

func init() {
	//rl.AddFileSystem(ASSETS)
}

func main() {
	game.RunGame()
}
