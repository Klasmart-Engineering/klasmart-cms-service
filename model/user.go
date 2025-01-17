package model

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/external"

	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/KL-Engineering/ro"

	"github.com/golang-jwt/jwt"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IUserModel interface {
	GetUserByAccount(ctx context.Context, account string) (*entity.User, error)
	RegisterUser(ctx context.Context, account string, password string, actType string) (*entity.User, error)
	UpdateAccountPassword(ctx context.Context, account string, password string) (*entity.User, error)
	ResetUserPassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
}

type UserModel struct{}

func (um *UserModel) GetUserByAccount(ctx context.Context, account string) (*entity.User, error) {
	return da.GetUserDA().GetUserByAccount(ctx, dbo.MustGetDB(ctx), account)
}

func (um *UserModel) RegisterUser(ctx context.Context, account string, password string, actType string) (*entity.User, error) {
	var user entity.User
	AmsID, err := external.GetUserServiceProvider().NewUser(ctx, nil, account)
	if err != nil {
		// Just warning now
		log.Warn(ctx, "RegisterUser:NewUser failed", log.String("account", account))
	}
	user.AmsID = AmsID
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, err := um.GetUserByAccount(ctx, account)
		if err == nil {
			log.Warn(ctx, "RegisterUser: already exist", log.String("account", account))
			return constant.ErrDuplicateRecord
		}
		if err != constant.ErrRecordNotFound {
			log.Error(ctx, "RegisterUser: GetUserByAccount failed", log.String("account", account), log.Err(err))
			return err
		}
		user.UserID = primitive.NewObjectID().Hex()
		if actType == constant.AccountPhone {
			user.Phone = account
		}
		if actType == constant.AccountEmail {
			user.Email = account
		}
		user.Salt, user.Secret = MakeSecretAndSalt(ctx, password)
		_, err = da.GetUserDA().InsertTx(ctx, tx, &user)
		if err != nil {
			log.Error(ctx, "RegisterUser: InsertTx failed", log.String("account", account), log.Err(err))
			return err
		}
		return nil
	})
	return &user, err
}

func (um *UserModel) UpdateAccountPassword(ctx context.Context, account string, password string) (user *entity.User, err error) {
	salt, secret := MakeSecretAndSalt(ctx, password)
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		var err error
		user, err = da.GetUserDA().GetUserByAccount(ctx, tx, account)
		if err != nil {
			log.Error(ctx, "UpdateAccountPassword: GetUserByAccount failed", log.String("account", account), log.Err(err))
			return err
		}
		user.Salt = salt
		user.Secret = secret
		user.UpdateAt = time.Now().Unix()
		_, err = da.GetUserDA().UpdateTx(ctx, tx, user)
		if err != nil {
			log.Error(ctx, "UpdateAccountPassword: UpdateTx failed", log.String("account", account), log.Err(err))
			return err
		}
		return nil
	})
	return
}

func (um *UserModel) ResetUserPassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		user, err := da.GetUserDA().GetUserByAccount(ctx, tx, userID)
		if err != nil {
			log.Error(ctx, "ResetUserPassword: GetUserByAccount failed", log.String("user_id", userID), log.Err(err))
			return err
		}
		pass := VerifySecretWithSalt(ctx, oldPassword, user.Secret, user.Salt)
		if !pass {
			log.Warn(ctx, "ResetUserPassword: not pass", log.String("user_id", userID))
			return constant.ErrUnAuthorized
		}
		user.Salt, user.Secret = MakeSecretAndSalt(ctx, newPassword)
		_, err = da.GetUserDA().UpdateTx(ctx, tx, user)
		if err != nil {
			log.Error(ctx, "ResetUserPassword:UpdateTx failed", log.String("user_id", userID), log.Err(err))
			return err
		}
		return nil
	})
	return err
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

type LoginClaim struct {
	jwt.StandardClaims
	Phone string `json:"phone"`
	ID    string `json:"id"`
}

func GetTokenFromUser(ctx context.Context, user *entity.User) (string, error) {
	now := time.Now()
	claim := &LoginClaim{
		StandardClaims: jwt.StandardClaims{
			Audience:  "Kidsloop",
			ExpiresAt: now.Add(time.Hour * 24 * constant.ValidDays).Unix(),
			IssuedAt:  now.Add(-30 * time.Second).Unix(),
			Issuer:    "Kidsloop_cn",
			NotBefore: 0,
			//Subject:   "authorization",
		},
		ID:    user.UserID,
		Phone: user.Phone,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS512, claim).SignedString(config.Get().KidsloopCNLoginConfig.PrivateKey)
	if err != nil {
		log.Error(ctx, "GetTokenFromUser:SignedString failed", log.Any("user", user))
		return "", err
	}
	return token, nil
}

func NewUserFromToken(token string) (*entity.User, error) {
	var claim LoginClaim
	_, err := jwt.ParseWithClaims(token, &claim, func(t *jwt.Token) (interface{}, error) {
		return config.Get().KidsloopCNLoginConfig.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return &entity.User{
		UserID: claim.ID,
	}, nil
}

func VerifyCode(ctx context.Context, codeKey string, code string) (bool, error) {
	client, err := ro.GetRedis(ctx)
	if err != nil {
		log.Error(ctx, "VerifyCode: GetRedis failed", log.String("code_key", codeKey), log.Err(err))
		return false, err
	}
	key := utils.GetHashKeyFromPlatformedString(codeKey)
	otpSecret, err := client.Get(ctx, key).Result()
	if err != nil && err.Error() == "redis: nil" {
		log.Error(ctx, "VerifyCode: redis nil", log.String("code_key", codeKey), log.String("key", key), log.Err(err))
		return false, constant.ErrUnAuthorized
	}
	if err != nil {
		log.Error(ctx, "VerifyCode: Get failed", log.String("code_key", codeKey), log.String("key", key), log.Err(err))
		return false, err
	}

	var authPassed bool
	baseSecret := OTPSecret(otpSecret)
	defer func() {
		if authPassed {
			log.Info(ctx, "VerifyCode: defer", log.Err(client.Del(ctx, key).Err()))
		}
	}()
	totp := baseSecret.getTOTPFromPool(codeKey)
	authPassed = totp.Verify(code)
	return authPassed, nil
}
func VerifySecretWithSalt(ctx context.Context, password string, secret string, salt string) bool {
	pwdHash := utils.Sha3Sign(strings.Replace(password, " ", "", -1), salt)
	if pwdHash == secret {
		return true
	}
	return false
}

func MakeSecretAndSalt(ctx context.Context, password string) (string, string) {
	salt := genSalt(ctx, "")
	secret := utils.Sha3Sign(strings.Replace(password, " ", "", -1), salt)
	return salt, secret
}
