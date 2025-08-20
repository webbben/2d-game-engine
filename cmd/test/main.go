package main

import (
	"log"

	"github.com/webbben/2d-game-engine/entity"
)

func main() {
	e := entity.Entity{}
	err := e.SaveJSON()
	if err != nil {
		log.Fatal(err)
	}
}
