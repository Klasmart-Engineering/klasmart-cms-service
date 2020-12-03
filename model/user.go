package model

import (
	"context"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/ro"

	"github.com/dgrijalva/jwt-go"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
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
	Co    string `json:"co"`
	Di    string `json:"di"`
	Email string `json:"em"`
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
			Subject:   "authorization",
		},
		ID:    user.UserID,
		Co:    "xx",
		Di:    "webpage",
		Email: user.Phone,
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
	otpSecret, err := client.Get(key).Result()
	if err != nil {
		log.Error(ctx, "VerifyCode: Get failed", log.String("code_key", codeKey), log.String("key", key), log.Err(err))
		return false, err
	}

	var authPassed bool
	baseSecret := OTPSecret(otpSecret)
	defer func() {
		if authPassed {
			log.Info(ctx, "VerifyCode: defer", log.Err(client.Del(key).Err()))
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
