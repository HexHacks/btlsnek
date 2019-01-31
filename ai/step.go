package ai

import (
	"github.com/battlesnakeio/starter-snake-go/api"
	"log"
)

func Step(req *api.SnakeRequest) string {
	step, err := NextMove(req)
	if err != nil {
		log.Print("Could not calculate next step:")
		log.Print(err)
		return "right"
	}
	return step
}
