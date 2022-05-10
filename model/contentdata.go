package model

import (
	"context"
	"errors"
	"strings"

	"github.com/KL-Engineering/kidsloop-cms-service/da"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type ContentData interface {
	Unmarshal(ctx context.Context, data string) error
	Marshal(ctx context.Context) (string, error)

	Validate(ctx context.Context, contentType entity.ContentType) error
	PrepareResult(ctx context.Context, tx *dbo.DBContext, content *entity.ContentInfo, operator *entity.Operator, ignorePermissionFilter bool) error
	PrepareSave(ctx context.Context, t entity.ExtraDataInRequest) error
	PrepareVersion(ctx context.Context) error
	SubContentIDs(ctx context.Context) []string

	ReplaceContentIDs(ctx context.Context, IDMap map[string]string)
}

func (cm *ContentModel) CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (ContentData, error) {
	var contentData ContentData
	switch contentType {
	case entity.ContentTypePlan:
		contentData = new(LessonData)
	case entity.ContentTypeMaterial:
		contentData = new(MaterialData)
	case entity.ContentTypeAssets:
		contentData = new(AssetsData)
	default:
		return nil, errors.New("unknown content type")
	}
	err := contentData.Unmarshal(ctx, data)
	if err != nil {
		return nil, err
	}
	return contentData, nil
}

func (c *ContentModel) ConvertContentObj(ctx context.Context, tx *dbo.DBContext, obj *entity.Content, operator *entity.Operator) (*entity.ContentInfo, error) {
	res, err := c.BatchConvertContentObj(ctx, tx, []*entity.Content{obj}, operator)
	if err != nil {
		return nil, err
	}
	return res[0], nil
}

func (c *ContentModel) ConvertContentObjWithProperties(ctx context.Context, obj *entity.Content, properties []*entity.ContentProperty) (*entity.ContentInfo, error) {
	contentData, err := c.CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		log.Error(ctx, "ConvertContentObjWithProperties: CreateContentData Failed",
			log.Err(err),
			log.Any("content", obj),
			log.Any("properties", properties))
		return nil, err
	}

	teacherManuals := make([]*entity.TeacherManualFile, 0)
	if obj.ContentType == entity.ContentTypePlan {
		planData := contentData.(*LessonData)
		if len(planData.TeacherManualBatch) > 0 {
			teacherManuals = planData.TeacherManualBatch
		} else if planData.TeacherManual != "" {
			teacherManuals = []*entity.TeacherManualFile{{
				ID:   planData.TeacherManual,
				Name: planData.TeacherManualName,
			}}
		}
	}

	subjects := make([]string, 0)
	developmentals := make([]string, 0)
	skills := make([]string, 0)
	ages := make([]string, 0)
	grades := make([]string, 0)
	keywords := make([]string, 0)
	outcomes := make([]string, 0)
	rejectReason := make([]string, 0)
	program := ""

	for i := range properties {
		switch properties[i].PropertyType {
		case entity.ContentPropertyTypeProgram:
			program = properties[i].PropertyID
		case entity.ContentPropertyTypeSubject:
			subjects = append(subjects, properties[i].PropertyID)
		case entity.ContentPropertyTypeCategory:
			developmentals = append(developmentals, properties[i].PropertyID)
		case entity.ContentPropertyTypeAge:
			ages = append(ages, properties[i].PropertyID)
		case entity.ContentPropertyTypeGrade:
			grades = append(grades, properties[i].PropertyID)
		case entity.ContentPropertyTypeSubCategory:
			skills = append(skills, properties[i].PropertyID)
		}
	}

	if obj.Keywords != "" {
		keywords = strings.Split(obj.Keywords, constant.StringArraySeparator)
	}
	if obj.Outcomes != "" {
		outcomes = strings.Split(obj.Outcomes, constant.StringArraySeparator)
	}
	if obj.RejectReason != "" {
		rejectReason = strings.Split(obj.RejectReason, constant.StringArraySeparator)
	}

	log.Info(ctx, "ConvertContentObjWithProperties: content properties",
		log.String("program", program),
		log.Strings("subjects", subjects),
		log.Strings("developmentals", developmentals),
		log.Strings("skills", skills),
		log.Strings("ages", ages),
		log.Strings("grades", grades))
	cm := &entity.ContentInfo{
		ID:                 obj.ID,
		ContentType:        obj.ContentType,
		Name:               obj.Name,
		Program:            program,
		Subject:            subjects,
		Category:           developmentals,
		SubCategory:        skills,
		Age:                ages,
		Grade:              grades,
		Keywords:           keywords,
		SourceType:         obj.SourceType,
		SuggestTime:        obj.SuggestTime,
		RejectReason:       rejectReason,
		Remark:             obj.Remark,
		Description:        obj.Description,
		Thumbnail:          obj.Thumbnail,
		Data:               obj.Data,
		Extra:              obj.Extra,
		Outcomes:           outcomes,
		Author:             obj.Author,
		TeacherManualBatch: teacherManuals,
		Creator:            obj.Creator,
		SelfStudy:          obj.SelfStudy.Bool(),
		DrawActivity:       obj.DrawActivity.Bool(),
		LessonType:         obj.LessonType,
		Org:                obj.Org,
		PublishStatus:      obj.PublishStatus,
		Version:            obj.Version,
		CreatedAt:          obj.CreateAt,
		UpdatedAt:          obj.UpdateAt,
		LatestID:           obj.LatestID,
	}

	return cm, nil
}
func (c *ContentModel) BatchConvertContentObj(ctx context.Context, tx *dbo.DBContext, objs []*entity.Content, operator *entity.Operator) ([]*entity.ContentInfo, error) {
	if len(objs) < 1 {
		return nil, nil
	}
	contentIDs := make([]string, len(objs))
	for i := range objs {
		contentIDs[i] = objs[i].ID
	}

	contentProperties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, tx, contentIDs)
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList",
			log.Err(err),
			log.Strings("ids", contentIDs))
		return nil, err
	}
	log.Info(ctx, "BatchGetByContentIDList result",
		log.Any("contentProperties", contentProperties))

	res := make([]*entity.ContentInfo, 0)
	for _, obj := range objs {
		log.Info(ctx, "Convert content object", log.String("extra", obj.Extra))
		contentData, err := c.CreateContentData(ctx, obj.ContentType, obj.Data)
		if err != nil {
			log.Error(ctx, "CreateContentData",
				log.Err(err),
				log.Any("obj", obj))
			return nil, err
		}
		teacherManuals := make([]*entity.TeacherManualFile, 0)
		if obj.ContentType == entity.ContentTypePlan {
			planData := contentData.(*LessonData)
			if len(planData.TeacherManualBatch) > 0 {
				teacherManuals = planData.TeacherManualBatch
			} else if planData.TeacherManual != "" {
				teacherManuals = []*entity.TeacherManualFile{{
					ID:   planData.TeacherManual,
					Name: planData.TeacherManualName,
				}}
			}
		}

		subjects := make([]string, 0)
		developmentals := make([]string, 0)
		skills := make([]string, 0)
		ages := make([]string, 0)
		grades := make([]string, 0)
		keywords := make([]string, 0)
		outcomes := make([]string, 0)
		rejectReason := make([]string, 0)
		program := ""

		for i := range contentProperties {
			if contentProperties[i].ContentID != obj.ID {
				continue
			}
			switch contentProperties[i].PropertyType {
			case entity.ContentPropertyTypeProgram:
				program = contentProperties[i].PropertyID
			case entity.ContentPropertyTypeSubject:
				subjects = append(subjects, contentProperties[i].PropertyID)
			case entity.ContentPropertyTypeCategory:
				developmentals = append(developmentals, contentProperties[i].PropertyID)
			case entity.ContentPropertyTypeAge:
				ages = append(ages, contentProperties[i].PropertyID)
			case entity.ContentPropertyTypeGrade:
				grades = append(grades, contentProperties[i].PropertyID)
			case entity.ContentPropertyTypeSubCategory:
				skills = append(skills, contentProperties[i].PropertyID)
			}
		}
		if obj.Keywords != "" {
			keywords = strings.Split(obj.Keywords, constant.StringArraySeparator)
		}
		if obj.Outcomes != "" {
			outcomes = strings.Split(obj.Outcomes, constant.StringArraySeparator)
		}
		if obj.RejectReason != "" {
			rejectReason = strings.Split(obj.RejectReason, constant.StringArraySeparator)
		}

		log.Info(ctx, "content properties",
			log.String("program", program),
			log.Strings("subjects", subjects),
			log.Strings("developmentals", developmentals),
			log.Strings("skills", skills),
			log.Strings("ages", ages),
			log.Strings("grades", grades))
		cm := &entity.ContentInfo{
			ID:                 obj.ID,
			ContentType:        obj.ContentType,
			Name:               obj.Name,
			Program:            program,
			Subject:            subjects,
			Category:           developmentals,
			SubCategory:        skills,
			Age:                ages,
			Grade:              grades,
			Keywords:           keywords,
			SourceType:         obj.SourceType,
			SuggestTime:        obj.SuggestTime,
			RejectReason:       rejectReason,
			Remark:             obj.Remark,
			Description:        obj.Description,
			Thumbnail:          obj.Thumbnail,
			Data:               obj.Data,
			Extra:              obj.Extra,
			Outcomes:           outcomes,
			Author:             obj.Author,
			TeacherManualBatch: teacherManuals,
			Creator:            obj.Creator,
			SelfStudy:          obj.SelfStudy.Bool(),
			DrawActivity:       obj.DrawActivity.Bool(),
			LessonType:         obj.LessonType,
			Org:                obj.Org,
			PublishStatus:      obj.PublishStatus,
			Version:            obj.Version,
			CreatedAt:          obj.CreateAt,
			UpdatedAt:          obj.UpdateAt,
			LatestID:           obj.LatestID,
		}
		res = append(res, cm)
	}

	return res, nil
}

func (c *ContentModel) BatchConvertContentObjForSearchContent(ctx context.Context, tx *dbo.DBContext, objs []*entity.Content, operator *entity.Operator) ([]*entity.ContentInfo, error) {
	if len(objs) < 1 {
		return nil, nil
	}
	contentIDs := make([]string, len(objs))
	for i := range objs {
		contentIDs[i] = objs[i].ID
	}

	res := make([]*entity.ContentInfo, 0)
	for _, obj := range objs {
		log.Info(ctx, "Convert content object", log.String("extra", obj.Extra))
		contentData, err := c.CreateContentData(ctx, obj.ContentType, obj.Data)
		if err != nil {
			log.Error(ctx, "CreateContentData",
				log.Err(err),
				log.Any("obj", obj))
			return nil, err
		}
		teacherManuals := make([]*entity.TeacherManualFile, 0)
		if obj.ContentType == entity.ContentTypePlan {
			planData := contentData.(*LessonData)
			if len(planData.TeacherManualBatch) > 0 {
				teacherManuals = planData.TeacherManualBatch
			} else if planData.TeacherManual != "" {
				teacherManuals = []*entity.TeacherManualFile{{
					ID:   planData.TeacherManual,
					Name: planData.TeacherManualName,
				}}
			}
		}

		keywords := make([]string, 0)
		outcomes := make([]string, 0)
		rejectReason := make([]string, 0)
		if obj.Keywords != "" {
			keywords = strings.Split(obj.Keywords, constant.StringArraySeparator)
		}
		if obj.Outcomes != "" {
			outcomes = strings.Split(obj.Outcomes, constant.StringArraySeparator)
		}
		if obj.RejectReason != "" {
			rejectReason = strings.Split(obj.RejectReason, constant.StringArraySeparator)
		}

		cm := &entity.ContentInfo{
			ID:                 obj.ID,
			ContentType:        obj.ContentType,
			Name:               obj.Name,
			Keywords:           keywords,
			SourceType:         obj.SourceType,
			SuggestTime:        obj.SuggestTime,
			RejectReason:       rejectReason,
			Remark:             obj.Remark,
			Description:        obj.Description,
			Thumbnail:          obj.Thumbnail,
			Data:               obj.Data,
			Extra:              obj.Extra,
			Outcomes:           outcomes,
			Author:             obj.Author,
			TeacherManualBatch: teacherManuals,
			Creator:            obj.Creator,
			SelfStudy:          obj.SelfStudy.Bool(),
			DrawActivity:       obj.DrawActivity.Bool(),
			LessonType:         obj.LessonType,
			Org:                obj.Org,
			PublishStatus:      obj.PublishStatus,
			Version:            obj.Version,
			CreatedAt:          obj.CreateAt,
			UpdatedAt:          obj.UpdateAt,
			LatestID:           obj.LatestID,
		}
		res = append(res, cm)
	}

	return res, nil
}
