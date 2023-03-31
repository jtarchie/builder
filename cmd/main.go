package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/jtarchie/builder"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("could not start logger: %s", err)
	}

	cli := &builder.CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err = ctx.Run(logger)
	ctx.FatalIfErrorf(err)
}
