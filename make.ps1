param (
	[string]$command="make"
)

function make {
	$env:GOOS="windows"
	$env:GOARCH="amd64"
	
	go build -o .\dist\game.exe
}

function make-wasm {
	$env:GOOS="js"
	$env:GOARCH="wasm"

	go build -o .\dist\wasm\main.wasm .
}

if ($command -eq "make") {
	make
} elseif ($command -eq "wasm") {
	make-wasm
} else {
	throw "invalid command"
}
