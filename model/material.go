package model

import (
	"context"
	"encoding/json"
	"errors"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var (
	ErrContentDataRequestSource = errors.New("material require source")
	ErrInvalidSourceExt         = errors.New("invalid source extension")

	ErrTeacherManual        = errors.New("teacher manual resource is not exist")
	ErrTeacherManualExtension        = errors.New("teacher manual extension is not supported")
	ErrInvalidTeacherManual = errors.New("invalid teacher manual")

	ErrTeacherManualBatchNameNil        = errors.New("teacher manual batch names is null")
	ErrTeacherManualNameNil        = errors.New("teacher manual names is null")
)

func NewMaterialData() *MaterialData {
	return &MaterialData{}
}

type MaterialData struct {
	InputSource int             `json:"input_source"`
	FileType    entity.FileType `json:"file_type"`
	Source      SourceID        `json:"source,omitempty"`
	Content     string          `json:"content,omitempty"`
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
	if this.Source.IsNil() && this.Content == "" {
		log.Error(ctx, "marshal material failed",
			log.String("source", string(this.Source)),
			log.Any("content", this.Content))
		return ErrContentDataRequestSource
	}

	switch this.InputSource {
	case entity.MaterialInputSourceH5p:
	case entity.MaterialInputSourceAssets:
		//查看assets?
		fallthrough
	case entity.MaterialInputSourceDisk:
		ext := this.Source.Ext()
		if !isArray(ext, constant.MaterialsExtension) {
			return ErrInvalidSourceExt
		}
	default:
		return ErrInvalidMaterialType
	}
	return nil
}

func (h *MaterialData) PrepareSave(ctx context.Context, t entity.ExtraDataInRequest) error {
	if h.InputSource == entity.MaterialInputSourceH5p {
		if h.Source == "" && h.Content != "" {
			h.FileType = entity.FileTypeH5pExtend
		} else if h.Source != "" && h.Content == "" {
			h.FileType = entity.FileTypeH5p
		} else {
			return ErrInvalidSourceOrContent
		}
		return nil
	}
	fileType, err := ExtensionToFileType(ctx, h.Source)
	if err != nil {
		return err
	}
	h.FileType = fileType
	return nil
}
func (h *MaterialData) SubContentIDs(ctx context.Context) []string {
	return nil
}
func (h *MaterialData) PrepareVersion(ctx context.Context) error {
	return nil
}
func (l *MaterialData) ReplaceContentIDs(ctx context.Context, IDMap map[string]string) {
}
func (h *MaterialData) PrepareResult(ctx context.Context, content *entity.ContentInfo, operator *entity.Operator) error {
	return nil
}
