package main

import (
	"github.com/alecthomas/kong"
	"github.com/jtarchie/builder"
)

func main() {
	cli := &builder.CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
