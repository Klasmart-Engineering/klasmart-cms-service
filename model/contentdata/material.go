package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

var(
	ErrInvalidContentType = errors.New("invalid content type")
	ErrContentDataRequestSource = errors.New("material require source")
	ErrInvalidMaterialType = errors.New("invalid material type")
	ErrInvalidSourceExt = errors.New("invalid source extension")

	ErrTeacherManual = errors.New("teacher manual resource is not exist")
	ErrInvalidTeacherManual = errors.New("invalid teacher manual")
)

func NewMaterialData() *MaterialData {
	return &MaterialData{}
}

type MaterialData struct {
	InputSource int      `json:"input_source"`
	FileType    entity.FileType      `json:"file_type"`
	Source      SourceID `json:"source"`
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
	if this.Source.IsNil() {
		log.Error(ctx, "marshal material failed", log.String("source", string(this.Source)))
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
		h.FileType = entity.FileTypeH5p
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
func (h *MaterialData) PrepareResult(ctx context.Context, operator *entity.Operator) error {
	return nil
}
