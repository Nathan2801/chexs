param (
	[string]$command="make"
)

function make-windows {
	$env:GOOS="windows"
	$env:GOARCH="amd64"
	
	cp .\go.default.mod .\go.mod
	go mod tidy
	go build -o dist .
}

function make-wasm {
	$env:GOOS="js"
	$env:GOARCH="wasm"

	cp .\go.wasm.mod .\go.mod
	go mod tidy
	go build -o dist\wasm\main.wasm .
}

if ($command -eq "make") {
	make-windows
} elseif ($command -eq "wasm") {
	make-wasm
} else {
	throw "invalid command"
}
