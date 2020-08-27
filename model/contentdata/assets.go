package contentdata

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
func (a *AssetsData) SubContentIds(ctx context.Context) ([]string ,error){
	return nil, nil
}

func (a *AssetsData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if strings.TrimSpace(a.Source) == "" {
		log.Error(ctx, "marshal material failed", log.String("source", a.Source))
		return errors.New("assets: require source")
	}
	return nil
}

func (h *AssetsData) PrepareResult(ctx context.Context) error {
	return nil
}

