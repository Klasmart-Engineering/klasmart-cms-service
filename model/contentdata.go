package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cm *ContentModel) CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (entity.ContentData, error) {
	var contentData entity.ContentData
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

func (c *ContentModel) ConvertContentObj(ctx context.Context, obj *entity.Content, operator *entity.Operator) (*entity.ContentInfo, error) {
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

	contentProperties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, dbo.MustGetDB(ctx), []string{obj.ID})
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList",
			log.Err(err),
			log.String("id", obj.ID))
		return nil, err
	}
	for i := range contentProperties {
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
		//PublishScope:  visibilitySettings,
		PublishStatus: obj.PublishStatus,
		Version:       obj.Version,
		CreatedAt:     obj.CreateAt,
		UpdatedAt:     obj.UpdateAt,
	}

	return cm, nil
}

func (c *ContentModel) BatchConvertContentObj(ctx context.Context, objs []*entity.Content, operator *entity.Operator) ([]*entity.ContentInfo, error) {
	if len(objs) < 1 {
		return nil, nil
	}
	contentIDs := make([]string, len(objs))
	for i := range objs {
		contentIDs[i] = objs[i].ID
	}

	contentProperties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, dbo.MustGetDB(ctx), contentIDs)
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList",
			log.Err(err),
			log.Strings("ids", contentIDs))
		return nil, err
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
		}
		res = append(res, cm)
	}

	return res, nil
}
