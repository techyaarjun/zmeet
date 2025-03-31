package main

import (
	_ "github.com/joho/godotenv/autoload"
	"simple-sfu/pkg/sfu"
)

func main() {
	s := sfu.NewServer()
	s.Start()
}
