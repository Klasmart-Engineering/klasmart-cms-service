package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

var(
	ErrInvalidContentType = errors.New("invalid content type")
	ErrContentDataRequestSource = errors.New("material require source")
	ErrInvalidMaterialInLesson = errors.New("invalid material in lesson")
	ErrInvalidMaterialType = errors.New("invalid material type")
	ErrInvalidSourceExt = errors.New("invalid source extension")
)

func NewMaterialData() *MaterialData {
	return &MaterialData{}
}

type MaterialData struct {
	InputSource int `json:"input_source"`
	FileType int `json:"file_type"`
	Source      string `json:"source"`
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

	switch this.InputSource {
	case entity.MaterialInputSourceH5p:
	case entity.MaterialInputSourceAssets:
		//查看assets?
		fallthrough
	case entity.MaterialInputSourceDisk:
		parts := strings.Split(this.Source, ".")
		if len(parts) < 2 {
			return errors.New("invalid source")
		}
		ext := parts[len(parts) - 1]
		ext = strings.ToLower(ext)
		if isArray(ext, constant.MaterialsExtension) {
			return errors.New("invalid source extension")
		}
	default:
		return ErrInvalidMaterialType
	}
	return nil
}

func (h *MaterialData) PrepareSave(ctx context.Context) error {
	if h.InputSource == entity.MaterialInputSourceH5p {
		return nil
	}
	fileType, err := ExtensionToFileType(ctx, h.Source)
	if err != nil{
		return err
	}
	h.FileType = fileType
	return nil
}
func (h *MaterialData) SubContentIds(ctx context.Context) []string{
	return nil
}
func (h *MaterialData) PrepareResult(ctx context.Context) error {
	return nil
}
