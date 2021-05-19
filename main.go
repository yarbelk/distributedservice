package main

import (
	"fmt"
	"time"

	"gitlab.com/yarbelk/grpcstuff/protostuff"
)

type ProtoStuffs struct {
	LastAction string
	Clock      time.Time
	log        []*protostuff.PlayerEventLog
}

func main() {
	fmt.Println("vim-go")
}
