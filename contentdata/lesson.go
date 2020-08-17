package contentdata

import (
	"context"
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
	panic("implement me")
}
func (l *LessonData) Marshal(ctx context.Context) (string, error) {
	panic("implement me")
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
