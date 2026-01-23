package main

import (
	"musiccat/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
