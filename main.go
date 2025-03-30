package main

import (
	"simple-sfu/pkg/sfu"
)

func main() {
	s := sfu.NewServer()
	s.Start()
}
