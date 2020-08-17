package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
)

func NewMaterialData() *MaterialData {
	return &MaterialData{}
}

type MaterialData struct {
	Source   string `json:"source"`
}
func (this *MaterialData) Unmarshal(ctx context.Context, data string) error {
	ins := MaterialData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		logger.WithContext(ctx).
			WithError(err).
			WithField("data", data).
			Errorln("material: unmarshal failed")
		return err
	}
	*this = ins
	return nil
}

func (this *MaterialData) Marshal(ctx context.Context) (string, error) {
	data, err := json.Marshal(this)
	if err != nil {
		logger.WithContext(ctx).
			WithError(err).
			WithField("this", this).
			Errorln("material: marshal failed")
		return "", err
	}
	return string(data), nil
}

func (this *MaterialData) Validate(ctx context.Context,  contentType int, tx *dbo.DBContext) error {
	if strings.TrimSpace(this.Source) == "" {
		logger.WithContext(ctx).
			WithField("source", this.Source).
			Errorln("material validate: require source")
		return errors.New("material: require source")
	}
	return nil
}

func (h *MaterialData) PrepareResult(ctx context.Context) error {
	return nil
}

func( h*MaterialData) RelatedContentIds(ctx context.Context) []int64 {
	return nil
}