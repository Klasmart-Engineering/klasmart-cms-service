package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

var(
	ErrInvalidContentType = errors.New("invalid content type")
	ErrContentDataRequestSource = errors.New("material require source")
	ErrInvalidMaterialInLesson = errors.New("invalid material in lesson")
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
		log.Error(ctx, "unmarshal material failed", log.String("data", data), log.Err(err))
		return err
	}
	*this = ins
	return nil
}

func (this *MaterialData) Marshal(ctx context.Context) (string, error) {
	data, err := json.Marshal(this)
	if err != nil {
		log.Error(ctx, "marshal material failed", log.Err(err))
		return "", err
	}
	return string(data), nil
}

func (this *MaterialData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if contentType != entity.ContentTypeMaterial {
		return ErrInvalidContentType
	}
	if strings.TrimSpace(this.Source) == "" {
		log.Error(ctx, "marshal material failed", log.String("source", this.Source))
		return ErrContentDataRequestSource
	}

	return nil
}
func (h *MaterialData) SubContentIds(ctx context.Context) []string{
	return nil
}
func (h *MaterialData) PrepareResult(ctx context.Context) error {
	return nil
}
