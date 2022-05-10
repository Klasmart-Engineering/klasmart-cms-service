package mapping

import (
	"fmt"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/urfave/cli/v2"
)

func NewMappingCommand() *cli.Command {
	return &cli.Command{
		Name:  "mapping",
		Usage: "academic profile mapping",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "mysql",
				Required: true,
				Usage:    "\033[1;33mRequired!\033[0m specify mysql `connection string`, eg: \"user:password@tcp(127.0.0.1:3306)/kidsloop2?parseTime=true&charset=utf8mb4\"",
			},
		},
		Action: func(c *cli.Context) error {
			mysql := c.String("mysql")

			log.Debug(c.Context, "mapping", log.String("mysql", mysql))

			if mysql == "" {
				return constant.ErrInvalidArgs
			}

			// set global config
			config.Set(&config.Config{
				DBConfig: config.DBConfig{
					ConnectionString: mysql,
					ShowLog:          true,
					ShowSQL:          true,
					MaxOpenConns:     16,
					MaxIdleConns:     2,
				},
			})

			// init dbo
			dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
				dbConf := config.Get().DBConfig
				c.ShowLog = dbConf.ShowLog
				c.ShowSQL = dbConf.ShowSQL
				c.MaxIdleConns = dbConf.MaxIdleConns
				c.MaxOpenConns = dbConf.MaxOpenConns
				c.ConnectionString = dbConf.ConnectionString
			})
			if err != nil {
				log.Error(c.Context, "create dbo failed", log.Err(err))
				return err
			}
			dbo.ReplaceGlobal(dboHandler)

			fmt.Println("start mapping")
			mapper := NewMapperImpl()
			services := GetServices()
			for _, service := range services {
				err := service.Do(c.Context, c, mapper)
				if err != nil {
					return err
				}
			}
			fmt.Println("Done")

			return nil
		},
	}
}
