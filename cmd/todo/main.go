package main

import (
	"os"

	v1 "github.com/AdventurerAmer/todo-api/cmd/todo/v1"
)

func main() {
	os.Exit(v1.Run())
}
