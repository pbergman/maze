package main

import (
	"log"
	"net/http"
	"strconv"

	_ "github.com/pbergman/maze/http"
	"github.com/pbergman/maze/config"
	"github.com/pbergman/maze/cli"
)

func main() {

	if !config.Config.Server {
		cli.App()
	} else {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Config.Port), nil))
	}
}

