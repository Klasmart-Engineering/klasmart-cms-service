package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IUserDA interface {
	GetUserByAccount(ctx context.Context, tx *dbo.DBContext, account string) (*entity.User, error)
	InsertTx(context.Context, *dbo.DBContext, *entity.User) (*entity.User, error)
	UpdateTx(context.Context, *dbo.DBContext, *entity.User) (int64, error)
}

var _userOnce sync.Once

var userDA *UserSqlDA

func GetUserDA() IUserDA {
	_userOnce.Do(func() {
		userDA = new(UserSqlDA)
	})
	return userDA
}
