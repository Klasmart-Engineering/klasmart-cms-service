package contentdata

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

func CreateContentData(ctx context.Context, fileType int, data string) (entity.ContentData, error) {
	return nil, nil
}

func ConvertContentObj(ctx context.Context, obj *entity.Content) (*entity.ContentInfo, error) {
	logger.WithContext(ctx).WithField("subject", "content").Infof("Convert content object, extra: %v", obj.Extra)

	contentData, err := CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		return nil, err
	}

	cm := &entity.ContentInfo{
		ID:            obj.ID,
		ContentType:   obj.ContentType,
		Name:          obj.Name,
		Program:       obj.Program,
		Subject:       obj.Subject,
		Developmental: obj.Developmental,
		Skills:        obj.Skills,
		Age:           obj.Age,
		Keywords:      strings.Split(obj.Keywords, ","),
		Description:   obj.Description,
		Thumbnail:     obj.Thumbnail,
		Data:          contentData,
		Extra:         obj.Data,
		Author:        obj.Author,
		AuthorName:    obj.AuthorName,
		Org:           obj.Org,
		PublishScope:  obj.PublishScope,
		PublishStatus: obj.PublishStatus,
		Version:       obj.Version,
	}

	return cm, nil
}