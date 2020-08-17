package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
	"time"
)

func (cm ContentModel) prepareCreateContentParams(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (*entity.Content, error) {
	now := time.Now()

	publishStatus := entity.NewContentPublishStatus(entity.ContentStatusDraft)

	if c.Data == nil {
		return nil, ErrNoContentData
	}
	err := c.Data.Validate(ctx, c.ContentType, dbo.MustGetDB(ctx))
	if err != nil{
		return nil, err
	}
	dataJSON, err := c.Data.Marshal(ctx)
	if err != nil{
		return nil, err
	}

	//获取publishScope&authorName
	publishScope := operator.OrgID
	authorName := operator.UserID

	return &entity.Content{
		ID:            utils.NewID(),
		ContentType:   c.ContentType,
		Name:          c.Name,
		Program:       c.Program,
		Subject:       c.Subject,
		Developmental: c.Developmental,
		Skills:        c.Skills,
		Age:           c.Age,
		Keywords:      strings.Join(c.Keywords, ","),
		Description:   c.Description,
		Thumbnail:     c.Thumbnail,
		Data:          dataJSON,
		Extra:         c.Extra,
		Author:        operator.UserID,
		AuthorName:    authorName,
		Org:           operator.OrgID,
		PublishScope:  publishScope,
		PublishStatus: publishStatus,
		Version:       0,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}, nil
}


func (cm ContentModel) prepareUpdateContentParams(ctx context.Context, content *entity.Content, data *entity.CreateContentRequest) (*entity.Content, error) {
	if data.Name != "" {
		content.Name = data.Name
	}
	if data.ContentType > 0 && data.Data != nil{
		content.ContentType = data.ContentType
	}
	if data.Program != "" {
		content.Program = data.Program
	}
	if data.Subject != "" {
		content.Subject = data.Subject
	}
	if data.Developmental != "" {
		content.Developmental = data.Developmental
	}
	if data.Skills != "" {
		content.Skills = data.Skills
	}
	if data.Age != "" {
		content.Age = data.Age
	}
	if data.Description != "" {
		content.Description = data.Description
	}
	if data.Thumbnail != "" {
		content.Thumbnail = data.Thumbnail
	}
	if data.Extra != "" {
		content.Extra = data.Extra
	}
	if len(data.Keywords) > 0 {
		content.Keywords = strings.Join(data.Keywords, ",")
	}
	if data.Data != nil{
		data.Data.Marshal(ctx)
	}

	if content.PublishStatus == entity.ContentStatusRejected {
		content.PublishStatus = entity.ContentStatusPending
	}

	//检查data
	err := data.Data.Validate(ctx, content.ContentType, dbo.MustGetDB(ctx))
	if err != nil{
		return nil, err
	}
	dataJSON, err := data.Data.Marshal(ctx)
	if err != nil{
		return nil, err
	}
	content.Data = dataJSON

	return content, nil
}

func (cm ContentModel) prepareCloneContentParams(ctx context.Context, content *entity.Content, user *entity.Operator) *entity.Content {
	content.ID = ""
	content.Author = user.UserID
	content.Org = user.OrgID
	content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusDraft)
	return content
}


func (cm ContentModel) prepareDeleteContentParams(ctx context.Context, content *entity.Content, publishStatus entity.ContentPublishStatus) (*entity.Content, error) {
	switch publishStatus {
	case entity.ContentStatusPublished:
		content.PublishStatus = entity.ContentStatusArchive
	//case entity.ContentStatusArchive:
	//	content.PublishStatus = entity.ContentStatusHidden
	default:
		now := time.Now()
		content.DeletedAt = &now
	}
	return content, nil
}