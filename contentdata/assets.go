package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"strings"
)

func NewAssetsData() *AssetsData {
	return &AssetsData{}
}

type AssetsData struct {
	Size 	 int 	`json:"size"`
	Source   string `json:"source"`
}

func (this *AssetsData) Unmarshal(ctx context.Context, data string) error {
	ins := AssetsData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		logger.WithContext(ctx).
			WithError(err).
			WithField("data", data).
			Errorln("assets: unmarshal failed")
		return err
	}
	*this = ins
	return nil
}

func (this *AssetsData) Marshal(ctx context.Context) (string, error) {
	//插入Size
	data, err := json.Marshal(this)
	if err != nil {
		logger.WithContext(ctx).
			WithError(err).
			WithField("this", this).
			Errorln("assets: marshal failed")
		return "", err
	}

	return string(data), nil
}

func (a *AssetsData) Validate(ctx context.Context,  contentType int, tx *dbo.DBContext) error {
	if strings.TrimSpace(a.Source) == "" {
		logger.WithContext(ctx).
			WithField("source", a.Source).
			Errorln("assets validate: require source")
		return errors.New("assets: require source")
	}
	return nil
}

func (h *AssetsData) PrepareResult(ctx context.Context) error {
	return nil
}

func(h *AssetsData) RelatedContentIds(ctx context.Context) []int64 {
	return nil
}
