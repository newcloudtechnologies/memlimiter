package main

import (
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/app"
)

func main() {
	logger := stdr.NewWithOptions(
		log.New(os.Stdout, "", log.LstdFlags),
		stdr.Options{LogCaller: stdr.All},
	)

	a := app.NewApp(logger, app.NewFactory(logger))
	a.Run()
}
