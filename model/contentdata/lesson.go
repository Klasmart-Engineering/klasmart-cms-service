package contentdata

import (
	"context"
	"encoding/json"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type LessonData struct {
	SegmentId  string          `json:"segmentId"`
	Condition  string          `json:"condition"`
	MaterialId string          `json:"materialId"`
	Material   *entity.Content `json:"material"`
	NextNode   []*LessonData   `json:"next"`
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

func (l *LessonData) lessonDataIterator(ctx context.Context, handleLessonData func(ctx context.Context, l *LessonData)) []string {
	arr := make([]string, 0)
	handleLessonData(ctx, l)
	for i := range l.NextNode {
		l.NextNode[i].lessonDataIterator(ctx, handleLessonData)
	}
	return arr
}

func (l *LessonData) lessonDataIteratorLoop(ctx context.Context, handleLessonData func(ctx context.Context, l *LessonData)) {
	lessonDataList := make([]*LessonData, 0)
	lessonDataList = append(lessonDataList, l)
	for len(lessonDataList) > 0 {
		//enqueue
		lessonData := lessonDataList[0]
		lessonDataList = lessonDataList[1:]

		handleLessonData(ctx, lessonData)
		for i := range lessonData.NextNode {
			//enqueue
			lessonDataList = append(lessonDataList, lessonData.NextNode[i])
		}
	}
}

func (l *LessonData) SubContentIds(ctx context.Context) ([]string, error) {
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	return materialList, nil
}

func (l *LessonData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if contentType != entity.ContentTypeLesson {
		return ErrInvalidContentType
	}
	//检查material合法性
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	_, data, err := da.GetDyContentDA().SearchContent(ctx, &da.DyContentCondition{
		IDS: materialList,
	})
	if err != nil {
		return err
	}
	if len(data) != len(materialList) {
		return ErrInvalidMaterialInLesson
	}

	//暂不检查Condition
	return nil
}

func (l *LessonData) PrepareResult(ctx context.Context) error {
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	_, contentList, err := da.GetDyContentDA().SearchContent(ctx, &da.DyContentCondition{
		IDS: materialList,
	})
	if err != nil {
		return err
	}

	contentMap := make(map[string]*entity.Content)
	for i := range contentList {
		contentMap[contentList[i].ID] = contentList[i]
	}
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		l.Material = contentMap[l.MaterialId]
	})
	return nil
}
