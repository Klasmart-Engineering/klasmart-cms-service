package contentdata

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

func NewAssetsData() *AssetsData {
	return &AssetsData{}
}

type AssetsData struct {
	Size 	 int 	`json:"size"`
	FileType int `json:"file_type"`
	Source   string `json:"source"`
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
func (a *AssetsData) SubContentIds(ctx context.Context) []string{
	return nil
}

func (a *AssetsData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if strings.TrimSpace(a.Source) == "" {
		log.Error(ctx, "marshal material failed", log.String("source", a.Source))
		return ErrContentDataRequestSource
	}
	parts := strings.Split(a.Source, ".")
	if len(parts) < 2 {
		return ErrInvalidSourceExt
	}
	ext := parts[len(parts) - 1]
	flag := false
	ext = strings.ToLower(ext)
	switch contentType {
	case entity.ContentTypeAssetImage:
		flag = isArray(ext, constant.AssetsImageExtension)
	case entity.ContentTypeAssetDocument:
		flag = isArray(ext, constant.AssetsDocExtension)
	case entity.ContentTypeAssetAudio:
		flag = isArray(ext, constant.AssetsAudioExtension)
	case entity.ContentTypeAssetVideo:
		flag = isArray(ext, constant.AssetsVideoExtension)
	}
	if !flag {
		return ErrInvalidSourceExt
	}

	return nil
}
func (h *AssetsData) PrepareSave(ctx context.Context) error {
	fileType, err := ExtensionToFileType(ctx, h.Source)
	if err != nil{
		return err
	}
	h.FileType = fileType
	return nil
}

func (h *AssetsData) PrepareResult(ctx context.Context) error {
	return nil
}

func isArray(ext string, extensions []string) bool{
	for i := range extensions {
		if extensions[i] == ext {
			return true
		}
	}
	return false
}


func ExtensionToFileType(ctx context.Context, resourceId string) (int, error) {
	if strings.TrimSpace(resourceId) == "" {
		log.Error(ctx, "marshal material failed", log.String("source", resourceId))
		return -1, ErrContentDataRequestSource
	}
	parts := strings.Split(resourceId, ".")
	if len(parts) < 2 {
		return -1, ErrInvalidSourceExt
	}
	ext := parts[len(parts) - 1]
	ext = strings.ToLower(ext)

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