package model

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

func NewAssetsData() *AssetsData {
	return &AssetsData{}
}

type AssetsData struct {
	Size     int             `json:"size"`
	FileType entity.FileType `json:"file_type"`
	Source   SourceID        `json:"source"`
}

type SourceID string

func (s SourceID) ConvertToPath(ctx context.Context) (string, error) {
	source := string(s)
	parts := strings.Split(source, "-")
	if len(parts) != 2 {
		log.Error(ctx, "ConvertToPath", log.String("resourceId", source))
		return "", constant.ErrInvalidArgs
	}
	return strings.Replace(source, "-", "/", -1), nil
}

func (s SourceID) Ext() string {
	parts := strings.Split(string(s), ".")
	if len(parts) < 2 {
		return ""
	}
	ext := parts[len(parts)-1]
	ext = strings.ToLower(ext)
	return ext
}
func (s SourceID) IsNil() bool {
	if strings.TrimSpace(string(s)) == "" {
		return true
	}
	return false
}

func (this *AssetsData) Unmarshal(ctx context.Context, data string) error {
	ins := AssetsData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		log.Error(ctx, "unmarshal material failed", log.String("data", data), log.Err(err))
		return err
	}
	*this = ins
	return nil
}

func (this *AssetsData) Marshal(ctx context.Context) (string, error) {
	//插入Size
	data, err := json.Marshal(this)
	if err != nil {
		log.Error(ctx, "marshal material failed", log.Err(err))
		return "", err
	}

	return string(data), nil
}
func (a *AssetsData) SubContentIDs(ctx context.Context) []string {
	return nil
}

func (a *AssetsData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if a.Source.IsNil() {
		log.Error(ctx, "marshal material failed", log.String("source", string(a.Source)))
		return ErrContentDataRequestSource
	}

	ext := a.Source.Ext()
	if !isArray(ext, constant.MaterialsExtension) {
		return ErrInvalidSourceExt
	}

	return nil
}
func (h *AssetsData) PrepareSave(ctx context.Context, t entity.ExtraDataInRequest) error {
	fileType, err := ExtensionToFileType(ctx, h.Source)
	if err != nil {
		return err
	}
	h.FileType = fileType
	return nil
}

func (h *AssetsData) PrepareVersion(ctx context.Context) error {
	return nil
}

func (h *AssetsData) PrepareResult(ctx context.Context, tx *dbo.DBContext, content *entity.ContentInfo, operator *entity.Operator, ignorePermissionFilter bool) error {
	return nil
}
func (l *AssetsData) ReplaceContentIDs(ctx context.Context, IDMap map[string]string) {
}
func isArray(ext string, extensions []string) bool {
	for i := range extensions {
		if extensions[i] == ext {
			return true
		}
	}
	return false
}

func ExtensionToFileType(ctx context.Context, sourceId SourceID) (entity.FileType, error) {
	if sourceId.IsNil() {
		log.Error(ctx, "marshal material failed", log.String("source", string(sourceId)))
		return -1, ErrContentDataRequestSource
	}
	ext := sourceId.Ext()
	if isArray(ext, constant.AssetsImageExtension) {
		return entity.FileTypeImage, nil
	}
	if isArray(ext, constant.AssetsDocExtension) {
		return entity.FileTypeDocument, nil
	}
	if isArray(ext, constant.AssetsAudioExtension) {
		return entity.FileTypeAudio, nil
	}
	if isArray(ext, constant.AssetsVideoExtension) {
		return entity.FileTypeVideo, nil
	}
	return -1, ErrInvalidSourceExt
}
