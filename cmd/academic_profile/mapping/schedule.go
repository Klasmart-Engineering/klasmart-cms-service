package mapping

import (
	"context"

	"github.com/urfave/cli/v2"
)

type Schedule struct {
}

func (s Schedule) Do(ctx context.Context, cliContext *cli.Context, mapper Mapper) error {
	// nothing todo, all schedule are stored using new program and subject id
	return nil
}
