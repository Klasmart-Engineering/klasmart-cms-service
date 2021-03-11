package model

import (
	"context"
	"encoding/json"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
)

type LessonData struct {
	SegmentId         string              `json:"segmentId"`
	Condition         string              `json:"condition"`
	MaterialId        string              `json:"materialId"`
	Material          *entity.ContentInfo `json:"material"`
	NextNode          []*LessonData       `json:"next"`
	TeacherManual     string              `json:"teacher_manual"`
	TeacherManualName string              `json:"teacher_manual_name"`

	TeacherManualBatch []string `json:"teacher_manual_batch"`
	TeacherManualBatchNames []string `json:"teacher_manual_batch_names"`
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

func (l *LessonData) LessonDataIteratorLoop(ctx context.Context, handleLessonData func(ctx context.Context, l *LessonData)) {
	l.lessonDataIteratorLoop(ctx, handleLessonData)
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
	h.TeacherManualBatch = t.TeacherManual
	h.TeacherManualBatchNames = t.TeacherManualName
	h.Material = nil
	return nil
}
func (l *LessonData) SubContentIDs(ctx context.Context) []string {
	materialList := make([]string, 0)
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		materialList = append(materialList, l.MaterialId)
	})
	return materialList
}

func (l *LessonData) ReplaceContentIDs(ctx context.Context, IDMap map[string]string) {
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		newID, ok := IDMap[l.MaterialId]
		if ok {
			l.MaterialId = newID
		}
	})
}

func (l2 *LessonData) checkTeacherManual(ctx context.Context, teacherManual string) error {
	teacherManualPairs := strings.Split(teacherManual, constant.TeacherManualSeparator)
	if len(teacherManualPairs) < 2 || teacherManualPairs[0] != string(storage.TeacherManualStoragePartition) {
		log.Warn(ctx, "teacher_manual is not exist in storage", log.String("TeacherManual", teacherManual),
			log.String("partition", string(storage.TeacherManualStoragePartition)))
		return ErrInvalidTeacherManual
	}

	_, exist := storage.DefaultStorage().ExistFile(ctx, storage.TeacherManualStoragePartition, teacherManualPairs[1])
	if !exist {
		log.Warn(ctx, "teacher_manual is not exist in storage", log.String("TeacherManual", teacherManual),
			log.String("partition", string(storage.TeacherManualStoragePartition)))
		return ErrTeacherManual
	}
	extensionPairs := strings.Split(teacherManual, ".")
	extension := extensionPairs[len(extensionPairs) - 1]
	ret := isArray(extension, constant.TeacherManualExtension)
	if !ret {
		log.Warn(ctx, "teacher_manual is extension is not supported",
			log.String("TeacherManual", teacherManual),
			log.String("extension", extension),
			log.Strings("extensionPairs", extensionPairs),
			log.Strings("expected", constant.TeacherManualExtension))
		return ErrTeacherManualExtension
	}
	return nil
}

func (l *LessonData) Validate(ctx context.Context, contentType entity.ContentType) error {
	if contentType != entity.ContentTypePlan {
		return ErrInvalidContentType
	}
	if l.TeacherManual != "" {
		if l.TeacherManualName == "" {
			log.Warn(ctx, "teacher_manual name is nil",
				log.String("TeacherManual", l.TeacherManual),
				log.String("TeacherManualName", l.TeacherManualName),
				log.Strings("TeacherManualBatch", l.TeacherManualBatch),
				log.Strings("TeacherManualBatchNames", l.TeacherManualBatchNames))
			return ErrTeacherManualNameNil
		}
		err := l.checkTeacherManual(ctx, l.TeacherManual)
		if err != nil{
			return err
		}
	}
	if(len(l.TeacherManualBatch) != len(l.TeacherManualBatchNames)) {
		log.Warn(ctx, "teacher_manual batch name is nil",
			log.String("TeacherManual", l.TeacherManual),
			log.String("TeacherManualName", l.TeacherManualName),
			log.Strings("TeacherManualBatch", l.TeacherManualBatch),
			log.Strings("TeacherManualBatchNames", l.TeacherManualBatchNames))
		return ErrTeacherManualBatchNameNil
	}
	for i := range l.TeacherManualBatch {
		err := l.checkTeacherManual(ctx, l.TeacherManualBatchNames[i])
		if err != nil{
			return err
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

func (l *LessonData) PrepareVersion(ctx context.Context) error {
	//list all related content ids
	ids := l.SubContentIDs(ctx)
	_, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
		IDS: ids,
	})
	if err != nil {
		return err
	}

	//build version map
	contentIDMap := make(map[string]string)
	for i := range contentList {
		if contentList[i].LatestID != "" {
			contentIDMap[contentList[i].ID] = contentList[i].LatestID
		} else {
			contentIDMap[contentList[i].ID] = contentList[i].ID
		}
	}

	//update materials to new version
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		newID, ok := contentIDMap[l.MaterialId]
		if ok {
			l.MaterialId = newID
		}
	})

	return nil
}

func (l *LessonData) isOrganizationHeadquarters(ctx context.Context, orgID string) (bool, error) {
	orgInfo, err := GetOrganizationPropertyModel().GetOrDefault(ctx, orgID)
	if err != nil {
		log.Warn(ctx, "parse get folder shared records params failed",
			log.Err(err),
			log.String("orgID", orgID))
		return false, err
	}
	if orgInfo.Type != entity.OrganizationTypeHeadquarters {
		log.Info(ctx, "org is not in head quarter",
			log.Any("orgInfo", orgInfo))
		return false, nil
	}
	return true, nil
}

func (l *LessonData) PrepareResult(ctx context.Context, content *entity.ContentInfo, operator *entity.Operator) error {
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

	isHeadQuarter, err := l.isOrganizationHeadquarters(ctx, content.Org)
	if err != nil {
		return err
	}
	if !isHeadQuarter {
		//if is not head quarter, filter unauthed materials
		contentList, err = l.filterMaterialsByPermission(ctx, contentList, operator)
		if err != nil {
			return err
		}
	} else {
		//if is head quarter, remove unpublished materials
		newContentList := make([]*entity.Content, 0)
		for i := range contentList {
			if contentList[i].PublishStatus == entity.ContentStatusPublished {
				newContentList = append(newContentList, contentList[i])
			}
		}
		contentList = newContentList
	}

	contentMap := make(map[string]*entity.Content)
	for i := range contentList {
		contentMap[contentList[i].ID] = contentList[i]
	}
	l.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		data, ok := contentMap[l.MaterialId]
		if ok {
			material, _ := ConvertContentObj(ctx, data, operator)
			l.Material = material
		}
	})
	return nil
}

func (l *LessonData) filterMaterialsByPermission(ctx context.Context, contentList []*entity.Content, operator *entity.Operator) ([]*entity.Content, error) {
	result := make([]*entity.Content, 0)
	pendingCheckAuthContents := make([]*entity.Content, 0)
	pendingCheckAuthIDs := make([]string, 0)
	for i := range contentList {
		if contentList[i].Org == operator.OrgID {
			result = append(result, contentList[i])
		} else {
			pendingCheckAuthContents = append(pendingCheckAuthContents, contentList[i])
			pendingCheckAuthIDs = append(pendingCheckAuthIDs, contentList[i].ID)
		}
	}
	//if contains materials is not from org, check auth
	if len(pendingCheckAuthContents) > 0 {
		condition := da.AuthedContentCondition{
			OrgIDs:     []string{operator.OrgID, constant.ShareToAll},
			ContentIDs: pendingCheckAuthIDs,
		}
		authRecords, err := da.GetAuthedContentRecordsDA().QueryAuthedContentRecords(ctx, dbo.MustGetDB(ctx), condition)
		if err != nil {
			log.Error(ctx, "search auth content failed",
				log.Err(err),
				log.Any("condition", condition))
			return nil, err
		}
		for i := range pendingCheckAuthContents {
			for j := range authRecords {
				if pendingCheckAuthContents[i].ID == authRecords[j].ContentID &&
					pendingCheckAuthContents[i].PublishStatus == entity.ContentStatusPublished {
					result = append(result, pendingCheckAuthContents[i])
					break
				}
			}
		}
	}

	return result, nil
}
