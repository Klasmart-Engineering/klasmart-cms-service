package mapping

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

func NewMappingCommand() *cli.Command {
	return &cli.Command{
		Name:  "mapping",
		Usage: "academic profile mapping",
		// Aliases: []string{"m"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "mysql",
				Required: true,
				Usage:    "specify mysql `connection string`, eg: (user:password@tcp(127.0.0.1:3306)/kidsloop2?parseTime=true&charset=utf8mb4)",
			},
			// &cli.StringFlag{
			// 	Name:     "redis-host",
			// 	Value:    "127.0.0.1",
			// 	Required: false,
			// 	Usage:    "specify redis host `address`",
			// },
			// &cli.IntFlag{
			// 	Name:     "redis-port",
			// 	Value:    6379,
			// 	Required: false,
			// 	Usage:    "specify redis `port`",
			// },
			// &cli.StringFlag{
			// 	Name:     "redis-password",
			// 	Value:    "",
			// 	Required: false,
			// 	Usage:    "specify redis `password`",
			// },
			&cli.StringFlag{
				Name:     "ams",
				Required: true,
				Usage:    "specify AMS `address`, eg: (https://api.alpha.kidsloop.net/user/)",
			},
		},
		Action: func(c *cli.Context) error {
			mysql := c.String("mysql")
			// redisHost := c.String("redis-host")
			// redisPort := c.Int("redis-port")
			// redisPassword := c.String("redis-password")
			ams := c.String("ams")

			log.Debug(c.Context, "mapping",
				log.String("mysql", mysql),
				// log.String("redis host", redisHost),
				// log.Int("redis port", redisPort),
				// log.String("redis password", redisPassword),
				log.String("ams", ams))

			if mysql == "" || ams == "" {
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
				// RedisConfig: config.RedisConfig{
				// 	Host:     redisHost,
				// 	Port:     redisPort,
				// 	Password: redisPassword,
				// },
				AMS: config.AMSConfig{
					EndPoint: ams,
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

			// // init ro
			// ro.SetConfig(&redis.Options{
			// 	Addr:     fmt.Sprintf("%v:%v", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
			// 	Password: config.Get().RedisConfig.Password,
			// })

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
