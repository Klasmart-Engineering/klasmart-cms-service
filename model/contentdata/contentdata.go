package contentdata

import (
	"context"
	"errors"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (entity.ContentData, error) {
	var contentData entity.ContentData
	switch contentType {
	case entity.ContentTypeLesson:
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

func ConvertContentObj(ctx context.Context, obj *entity.Content) (*entity.ContentInfo, error) {
	log.Info(ctx, "Convert content object", log.String("extra", obj.Extra))
	//contentData, err := CreateContentData(ctx, obj.ContentType, obj.Data)
	//if err != nil {
	//	return nil, err
	//}
	programs := make([]string, 0)
	subjects := make([]string, 0)
	developmentals := make([]string, 0)
	skills := make([]string, 0)
	ages := make([]string, 0)
	grades := make([]string, 0)
	keywords := make([]string, 0)
	outcomes := make([]string, 0)
	rejectReason := make([]string, 0)
	if obj.Program != "" {
		programs = strings.Split(obj.Program, ",")
	}
	if obj.Subject != "" {
		subjects = strings.Split(obj.Subject, ",")
	}
	if obj.Developmental != "" {
		developmentals = strings.Split(obj.Developmental, ",")
	}
	if obj.Skills != "" {
		skills = strings.Split(obj.Skills, ",")
	}
	if obj.Age != "" {
		ages = strings.Split(obj.Age, ",")
	}
	if obj.Grade != "" {
		grades = strings.Split(obj.Grade, ",")
	}
	if obj.Keywords != "" {
		keywords = strings.Split(obj.Keywords, ",")
	}
	if obj.Outcomes != "" {
		outcomes = strings.Split(obj.Outcomes, ",")
	}
	if obj.RejectReason != "" {
		rejectReason = strings.Split(obj.RejectReason, ",")
	}

	cm := &entity.ContentInfo{
		ID:            obj.ID,
		ContentType:   obj.ContentType,
		Name:          obj.Name,
		Program:       programs,
		Subject:       subjects,
		Developmental: developmentals,
		Skills:        skills,
		Age:           ages,
		Grade:         grades,
		Keywords:      keywords,
		SuggestTime:   obj.SuggestTime,
		RejectReason:  rejectReason,
		Description:   obj.Description,
		Thumbnail:     obj.Thumbnail,
		Data:          obj.Data,
		Extra:         obj.Extra,
		Outcomes:      outcomes,
		Author:        obj.Author,
		AuthorName:    obj.AuthorName,
		SelfStudy: obj.SelfStudy.Bool(),
		DrawActivity: obj.DrawActivity.Bool(),
		LessonType: obj.LessonType,
		Org:           obj.Org,
		PublishScope:  obj.PublishScope,
		PublishStatus: obj.PublishStatus,
		Version:       obj.Version,
		CreatedAt:     obj.CreateAt,
		UpdatedAt:		obj.UpdateAt,
	}

	return cm, nil
}
