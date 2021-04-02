package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (entity.ContentData, error) {
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

func ConvertContentObj(ctx context.Context, obj *entity.Content, operator *entity.Operator) (*entity.ContentInfo, error) {
	log.Info(ctx, "Convert content object", log.String("extra", obj.Extra))
	contentData, err := CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		return nil, err
	}
	teacherManuals := make([]*entity.TeacherManualFile, 0)
	if obj.ContentType == entity.ContentTypePlan {
		planData := contentData.(*LessonData)
		if len(planData.TeacherManualBatch) > 0 {
			teacherManuals = planData.TeacherManualBatch
		}else if planData.TeacherManual != ""{
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
	if obj.Subject != "" {
		subjects = strings.Split(obj.Subject, constant.StringArraySeparator)
	}
	if obj.Developmental != "" {
		developmentals = strings.Split(obj.Developmental, constant.StringArraySeparator)
	}
	if obj.Skills != "" {
		skills = strings.Split(obj.Skills, constant.StringArraySeparator)
	}
	if obj.Age != "" {
		ages = strings.Split(obj.Age, constant.StringArraySeparator)
	}
	if obj.Grade != "" {
		grades = strings.Split(obj.Grade, constant.StringArraySeparator)
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
	//user, err := external.GetUserServiceProvider().Get(ctx, operator, obj.Author)
	//authorName := ""
	//if err != nil{
	//	log.Warn(ctx, "get user info failed", log.Err(err), log.Any("obj", obj))
	//}else{
	//	authorName = user.Name
	//}

	cm := &entity.ContentInfo{
		ID:            obj.ID,
		ContentType:   obj.ContentType,
		Name:          obj.Name,
		Program:       obj.Program,
		Subject:       subjects,
		Developmental: developmentals,
		Skills:        skills,
		Age:           ages,
		Grade:         grades,
		Keywords:      keywords,
		SourceType:		obj.SourceType,
		SuggestTime:   obj.SuggestTime,
		RejectReason:  rejectReason,
		Remark: 	   obj.Remark,
		Description:   obj.Description,
		Thumbnail:     obj.Thumbnail,
		Data:          obj.Data,
		Extra:         obj.Extra,
		Outcomes:      outcomes,
		Author:        obj.Author,
		TeacherManualBatch: teacherManuals,
		Creator:		obj.Creator,
		SelfStudy:     obj.SelfStudy.Bool(),
		DrawActivity:  obj.DrawActivity.Bool(),
		LessonType:    obj.LessonType,
		Org:           obj.Org,
		PublishScope:  obj.PublishScope,
		PublishStatus: obj.PublishStatus,
		Version:       obj.Version,
		CreatedAt:     obj.CreateAt,
		UpdatedAt:     obj.UpdateAt,
	}

	return cm, nil
}
