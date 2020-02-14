package main

import (
	"log"
	"math/rand"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/battlesnakeio/engine/cmd/engine/commands"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	rand.Seed(time.Now().UnixNano())
	commands.Execute()
}
