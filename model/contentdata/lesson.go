package contentdata

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
	"strings"
)

type LessonData struct {
	SegmentId  string          `json:"segmentId"`
	Condition  string          `json:"condition"`
	MaterialId string          `json:"materialId"`
	Material   *entity.ContentInfo `json:"material"`
	NextNode   []*LessonData   `json:"next"`
	TeacherManual string 	`json:"teacher_manual"`
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
func (h *LessonData) PrepareSave(ctx context.Context, t entity.ExtraDataInRequest) error {
	h.TeacherManual = t.TeacherManual
	return nil
}
func (l *LessonData) SubContentIds(ctx context.Context) []string {
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	return materialList
}

func (l *LessonData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if contentType != entity.ContentTypeLesson {
		return ErrInvalidContentType
	}
	if l.TeacherManual != "" {
		teacherManualPairs :=strings.Split(l.TeacherManual, "-")
		if len(teacherManualPairs) < 2 || teacherManualPairs[0] != string(storage.TeacherManualStoragePartition) {
			log.Warn(ctx, "teacher_manual is not exist in storage", log.String("TeacherManual", l.TeacherManual),
				log.String("partition", string(storage.TeacherManualStoragePartition)))
			return ErrInvalidTeacherManual
		}

		_, exist := storage.DefaultStorage().ExistFile(ctx, storage.TeacherManualStoragePartition, teacherManualPairs[1])
		if !exist {
			log.Warn(ctx, "teacher_manual is not exist in storage", log.String("TeacherManual", l.TeacherManual),
				log.String("partition", string(storage.TeacherManualStoragePartition)))
			return ErrTeacherManual
		}
	}

	//暂时不做检查
	//检查material合法性
	//materialList := make([]string, 0)
	//l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
	//	materialList = append(materialList, l.MaterialId)
	//})
	//_, data, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
	//	IDS: materialList,
	//})
	//if err != nil {
	//	return err
	//}
	////检查是否有不存在的material
	//for i := range materialList {
	//	flag := false
	//	for j := range data {
	//		if data[j].ID == materialList[i] {
	//			flag = true
	//			break
	//		}
	//	}
	//	if !flag {
	//		return ErrInvalidMaterialInLesson
	//	}
	//}

	//暂不检查Condition
	return nil
}

func (l *LessonData) PrepareResult(ctx context.Context, operator *entity.Operator) error {
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	_, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
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
		data, ok:= contentMap[l.MaterialId]
		if ok {
			material, _ := ConvertContentObj(ctx, data, operator)
			l.Material = material
		}
	})
	return nil
}
