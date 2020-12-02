package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbo "gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"
)

func (cm ContentModel) getSourceType(ctx context.Context, c entity.CreateContentRequest, d entity.ContentData) string {
	if c.ContentType == entity.ContentTypeLesson {
		return constant.SourceTypeLesson
	}
	if c.ContentType == entity.ContentTypeAssets {
		return constant.SourceTypeAssets
	}
	materialData := d.(*contentdata.MaterialData)
	return fmt.Sprintf(constant.SourceTypeMaterialPrefix + materialData.FileType.String())
}

func (cm ContentModel) checkSuggestTime(ctx context.Context, suggestTime int, contentType entity.ContentType, subIds []string) error {
	if contentType == entity.ContentTypeLesson {
		//if content type is lesson, check suggest time
		subContents, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), subIds)
		if err != nil {
			log.Error(ctx, "get content by id list failed", log.Err(err), log.Strings("subIds", subIds))
			return ErrReadContentFailed
		}
		timeTotal := 0
		for i := range subContents {
			timeTotal = timeTotal + subContents[i].SuggestTime
		}

		if suggestTime < timeTotal {
			log.Warn(ctx, "suggest time too small", log.Err(err),
				log.Int("timeTotal", timeTotal),
				log.Int("suggestTime", suggestTime),
				log.Any("subContents", subContents))
			return ErrSuggestTimeTooSmall
		}
	}
	return nil
}

func (cm ContentModel) prepareCreateContentParams(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (*entity.Content, error) {
	publishStatus := entity.NewContentPublishStatus(entity.ContentStatusDraft)

	if c.Data == "" {
		return nil, ErrNoContentData
	}
	cd, err := contentdata.CreateContentData(ctx, c.ContentType, c.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, ErrInvalidContentData
	}

	err = cd.Validate(ctx, c.ContentType)
	if err != nil {
		log.Warn(ctx, "validate content data failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, err
	}

	//check suggest time
	err = cm.checkSuggestTime(ctx, c.SuggestTime, c.ContentType, cd.SubContentIds(ctx))
	if err != nil {
		log.Warn(ctx, "check suggest time failed", log.Err(err), log.Any("req", c))
		return nil, err
	}

	err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{TeacherManual: c.TeacherManual, TeacherManualName: c.TeacherManualName})
	if err != nil {
		log.Warn(ctx, "prepare save content data failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, ErrInvalidContentData
	}

	data, err := cd.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "prepare save content data failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, ErrMarshalContentDataFailed
	}
	c.Data = data

	//get publishScope&authorName
	publishScope := c.PublishScope

	if c.SourceType == "" {
		c.SourceType = cm.getSourceType(ctx, c, cd)
	}

	//若为asset，直接发布
	if c.ContentType == entity.ContentTypeAssets {
		publishStatus = entity.NewContentPublishStatus(entity.ContentStatusPublished)
		c.SelfStudy = false
		c.DrawActivity = false
		c.LessonType = ""
	}
	if c.ContentType == entity.ContentTypeLesson {
		c.LessonType = ""
	}

	path := constant.FolderRootPath
	return &entity.Content{
		//ID:            utils.NewID(),
		ContentType:   c.ContentType,
		Name:          c.Name,
		Program:       c.Program,
		Subject:       strings.Join(c.Subject, ","),
		Developmental: strings.Join(c.Developmental, ","),
		Skills:        strings.Join(c.Skills, ","),
		Age:           strings.Join(c.Age, ","),
		Grade:         strings.Join(c.Grade, ","),
		Keywords:      strings.Join(c.Keywords, ","),
		Description:   c.Description,
		Thumbnail:     c.Thumbnail,
		SuggestTime:   c.SuggestTime,
		Data:          c.Data,
		Extra:         c.Extra,
		LessonType:    c.LessonType,
		DirPath:       path,
		SelfStudy:     c.SelfStudy.Int(),
		DrawActivity:  c.DrawActivity.Int(),
		Outcomes:      strings.Join(c.Outcomes, ","),
		Author:        operator.UserID,
		Creator:       operator.UserID,
		LockedBy:      constant.LockedByNoBody,
		Org:           operator.OrgID,
		PublishScope:  publishScope,
		PublishStatus: publishStatus,
		Version:       1,
	}, nil
}

func (cm ContentModel) prepareUpdateContentParams(ctx context.Context, content *entity.Content, data *entity.CreateContentRequest) (*entity.Content, error) {
	if data.Name != "" {
		content.Name = data.Name
	}
	if data.ContentType > 0 && data.Data != "" {
		content.ContentType = data.ContentType
	}
	if data.Program != "" {
		content.Program = data.Program
	}
	if data.Subject != nil {
		content.Subject = strings.Join(data.Subject, ",")
	}
	if data.Developmental != nil {
		content.Developmental = strings.Join(data.Developmental, ",")
	}
	if data.Skills != nil {
		content.Skills = strings.Join(data.Skills, ",")
	}
	if data.Age != nil {
		content.Age = strings.Join(data.Age, ",")
	}
	if data.Grade != nil {
		content.Grade = strings.Join(data.Grade, ",")
	}
	if data.Description != "" {
		content.Description = data.Description
	}
	if data.Thumbnail != "" {
		content.Thumbnail = data.Thumbnail
	}
	if data.Outcomes != nil {
		content.Outcomes = strings.Join(data.Outcomes, ",")
	}
	if data.Extra != "" {
		content.Extra = data.Extra
	}
	if len(data.Keywords) > 0 {
		content.Keywords = strings.Join(data.Keywords, ",")
	}
	if data.SuggestTime > 0 {
		content.SuggestTime = data.SuggestTime
	}
	if data.ContentType == entity.ContentTypeMaterial {
		content.DrawActivity = data.DrawActivity.Int()
		content.SelfStudy = data.SelfStudy.Int()
	}

	if data.ContentType == entity.ContentTypeMaterial && data.LessonType != "" {
		content.LessonType = data.LessonType
	}

	if content.PublishStatus == entity.ContentStatusRejected {
		content.PublishStatus = entity.ContentStatusDraft
	}
	//若已发布，不能修改publishScope
	if content.PublishStatus == entity.ContentStatusDraft ||
		content.PublishStatus == entity.ContentStatusRejected {
		content.PublishScope = data.PublishScope
	}

	//Asset修改后直接发布
	if content.ContentType.IsAsset() {
		content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusPublished)
	}

	if data.SourceType != "" {
		content.SourceType = data.SourceType
	}

	//检查data
	if data.Data != "" {
		cd, err := contentdata.CreateContentData(ctx, data.ContentType, data.Data)
		if err != nil {
			return nil, ErrInvalidContentData
		}
		err = cd.Validate(ctx, content.ContentType)
		if err != nil {
			return nil, err
		}

		err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{TeacherManual: data.TeacherManual, TeacherManualName: data.TeacherManualName})
		if err != nil {
			return nil, ErrInvalidContentData
		}

		d, err := cd.Marshal(ctx)
		if err != nil {
			return nil, ErrMarshalContentDataFailed
		}

		content.Data = d

		if data.SourceType == "" {
			data.SourceType = cm.getSourceType(ctx, *data, cd)
		}

		//check suggest time
		err = cm.checkSuggestTime(ctx, data.SuggestTime, data.ContentType, cd.SubContentIds(ctx))
		if err != nil {
			log.Warn(ctx, "check suggest time failed", log.Err(err), log.Any("req", data), log.Any("content", content))
			return nil, err
		}
	}
	content.UpdateAt = time.Now().Unix()

	return content, nil
}

func (cm ContentModel) prepareCloneContentParams(ctx context.Context, content *entity.Content, user *entity.Operator) *entity.Content {
	content.SourceID = content.ID
	content.Version = content.Version + 1
	content.ID = ""
	content.LockedBy = constant.LockedByNoBody
	content.Author = user.UserID
	//content.Author = user.UserID
	//content.Org = user.OrgID
	content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusDraft)
	return content
}

func (cm ContentModel) prepareDeleteContentParams(ctx context.Context, content *entity.Content, publishStatus entity.ContentPublishStatus, user *entity.Operator) *entity.Content {
	content.DirPath = constant.FolderRootPath

	//assets则隐藏
	if content.ContentType.IsAsset() {
		content.PublishStatus = entity.ContentStatusHidden
		return content
	}

	switch publishStatus {
	case entity.ContentStatusPublished:
		content.PublishStatus = entity.ContentStatusArchive
	case entity.ContentStatusArchive:
		fallthrough
	default:
		now := time.Now()
		content.DeleteAt = now.Unix()
	}
	return content
}

func (cm *ContentModel) preparePublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//若content为archive，则直接发布
	if content.PublishStatus == entity.ContentStatusArchive {
		content.PublishStatus = entity.ContentStatusPublished
		content.UpdateAt = time.Now().Unix()
		//更新content的path
		existPath, err := GetFolderModel().ExistsPath(ctx, tx, entity.OwnerTypeOrganization, entity.FolderItemTypeFolder, content.DirPath, entity.FolderPartitionMaterialAndPlans, user)
		if err != nil {
			log.Error(ctx, "search content folder failed",
				log.Err(err), log.Any("content", content))
			return err
		}
		if !existPath {
			content.DirPath = constant.FolderRootPath
		}
		return nil
	}

	err := cm.checkPublishContent(ctx, tx, content, user)
	if err != nil {
		log.Warn(ctx, "check content scope & sub content scope failed", log.Err(err))
		return err
	}

	content.PublishStatus = entity.ContentStatusPending
	content.UpdateAt = time.Now().Unix()
	return nil
}
