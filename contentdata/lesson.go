package contentdata

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type LessonData struct {
	SegmentId int `json:"segmentId"`
	Condition string `json:"condition"`
	MaterialId string `json:"materialId"`
	Material 	*entity.Content `json:"material"`
	NextNode 	[]LessonData `json:"next"`
}

func (l *LessonData) Unmarshal(ctx context.Context, data string) error {
	ins := LessonData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		log.Error(ctx, "unmarshal material failed", log.String("data", data), log.Err(err))
		return err
	}
	*l = ins
	return nil
}
func (l *LessonData) Marshal(ctx context.Context) (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		log.Error(ctx, "marshal material failed", log.Err(err))
		return "", err
	}
	return string(data), nil
}

//PrepareSave?

func (l *LessonData) Validate(ctx context.Context,  contentType int, tx *dbo.DBContext) error {
	panic("implement me")
}

func (l *LessonData) RelatedContentIds(ctx context.Context) []int64 {
	panic("implement me")
}

func (l *LessonData) PrepareResult(ctx context.Context) error {
	panic("implement me")
}
