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
	case entity.ContentTypeAssetVideo:
		fallthrough
	case entity.ContentTypeAssetImage:
		fallthrough
	case entity.ContentTypeAssetAudio:
		fallthrough
	case entity.ContentTypeAssetDocument:
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

	cm := &entity.ContentInfo{
		ID:            obj.ID,
		ContentType:   obj.ContentType,
		Name:          obj.Name,
		Program:       strings.Split(obj.Program, ","),
		Subject:       strings.Split(obj.Subject, ","),
		Developmental: strings.Split(obj.Developmental, ","),
		Skills:        strings.Split(obj.Skills, ","),
		Age:           strings.Split(obj.Age, ","),
		Grade:         strings.Split(obj.Grade, ","),
		Keywords:      strings.Split(obj.Keywords, ","),
		SuggestTime:   obj.SuggestTime,
		RejectReason:  obj.RejectReason,
		Description:   obj.Description,
		Thumbnail:     obj.Thumbnail,
		Data:          obj.Data,
		Extra:         obj.Data,
		Author:        obj.Author,
		AuthorName:    obj.AuthorName,
		Org:           obj.Org,
		PublishScope:  obj.PublishScope,
		PublishStatus: obj.PublishStatus,
		Version:       obj.Version,
		CreatedAt:     obj.CreateAt,
	}

	return cm, nil
}
