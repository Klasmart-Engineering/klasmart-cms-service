package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "ap",
		Usage:     "academic profile utils",
		UsageText: "apm command [command options] [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "s",
				Aliases: []string{"server"},
				Value:   "http://127.0.0.1:9100",
				Usage:   "set server address",
			},
		},
	}

	// for _, cmd := range command.Commands {
	// 	app.Commands = append(app.Commands, cmd.CliCommand())
	// }

	ctx := context.Background()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		fmt.Println("\033[31m" + err.Error() + "\033[0m ")
		return
	}
}
