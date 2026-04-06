package main

import (
	"embed"
	"os"

	v1 "github.com/AdventurerAmer/todo-api/cmd/todo/v1"
)

//go:embed templates
var templates embed.FS

func main() {
	os.Exit(v1.Run(templates))
}
