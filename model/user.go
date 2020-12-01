package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IUserModel interface {
	GetUserByAccount(ctx context.Context, account string) (*entity.User, error)
	RegisterUser(ctx context.Context, account string, password string, actType string) (*entity.User, error)
}

type UserModel struct{}

func (um *UserModel) GetUserByAccount(ctx context.Context, account string) (*entity.User, error) {
	return da.GetUserDA().GetUserByAccount(ctx, dbo.MustGetDB(ctx), account)
}

func (um *UserModel) RegisterUser(ctx context.Context, account string, password string, actType string) (*entity.User, error) {
	var user entity.User
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, err := um.GetUserByAccount(ctx, account)
		if err == nil {
			return constant.ErrDuplicateRecord
		}
		if err == constant.ErrRecordNotFound {
			user.UserID = primitive.NewObjectID().Hex()
			if actType == constant.AccountPhone {
				user.Phone = account
			}
			if actType == constant.AccountEmail {
				user.Email = account
			}
			user.Salt, user.Secret = MakeSecretAndSalt(ctx, password)
			_, err := da.GetUserDA().InsertTx(ctx, tx, &user)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	})
	return &user, err
}

var (
	_userModel     IUserModel
	_userModelOnce sync.Once
)

func GetUserModel() IUserModel {
	_userModelOnce.Do(func() {
		_userModel = new(UserModel)
	})
	return _userModel
}
