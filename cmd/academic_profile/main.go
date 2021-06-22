package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/academic_profile/mapping"
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
