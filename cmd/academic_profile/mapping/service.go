package mapping

import (
	"context"
	"sync"

	"github.com/urfave/cli/v2"
)

type Service interface {
	Do(ctx context.Context, cliContext *cli.Context, mapper Mapper) error
}

var (
	_serviceOnce sync.Once
	_services    = []Service{}
)

func GetServices() []Service {
	_serviceOnce.Do(func() {
		_services = []Service{
			&ContentService{},
			&OutcomeService{},
			&Schedule{},
		}
	})

	return _services
}
