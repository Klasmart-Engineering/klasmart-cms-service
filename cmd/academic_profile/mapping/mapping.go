package mapping

import (
	"github.com/urfave/cli/v2"
)

func NewMappingCommand() *cli.Command {
	return &cli.Command{
		Name:  "mapping",
		Usage: "academic profile mapping",
		// Aliases: []string{"m"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "mysql",
				Value:    "",
				Required: true,
				Usage:    "\033[1;33mRequired!\033[0m specify mysql connection string",
			},
			&cli.StringFlag{
				Name:     "redis",
				Value:    "",
				Required: true,
				Usage:    "\033[1;33mRequired!\033[0m specify redis `address`",
			},
		},
		Action: func(c *cli.Context) error {
			// arch := c.String("arch")
			// name := c.String("n")

			return nil
		},
	}
}
