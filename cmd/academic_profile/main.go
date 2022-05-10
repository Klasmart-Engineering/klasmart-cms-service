package main

import (
	"context"
	"fmt"
	"os"

	"github.com/KL-Engineering/kidsloop-cms-service/cmd/academic_profile/mapping"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "ap",
		Usage:     "academic profile utils",
		UsageText: "ap command [command options] [arguments...]",
		Commands: []*cli.Command{
			mapping.NewMappingCommand(),
		},
	}

	ctx := context.Background()
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		fmt.Println("\033[31m" + err.Error() + "\033[0m ")
		return
	}
}
