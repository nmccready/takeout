package main

import (
	"github.com/nmccready/takeout/src/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
