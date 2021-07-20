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
)

func (cm ContentModel) getSourceType(ctx context.Context, c entity.CreateContentRequest, d ContentData) string {
	if c.ContentType == entity.ContentTypePlan {
		return constant.SourceTypeLesson
	}
	if c.ContentType == entity.ContentTypeAssets {
		return constant.SourceTypeAssets
	}
	materialData := d.(*MaterialData)
	return fmt.Sprintf(constant.SourceTypeMaterialPrefix + materialData.FileType.String())
}

func (cm ContentModel) checkSuggestTime(ctx context.Context, suggestTime int, contentType entity.ContentType, subIds []string) error {
	if contentType == entity.ContentTypePlan {
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
	cd, err := cm.CreateContentData(ctx, c.ContentType, c.Data)
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
	err = cm.checkSuggestTime(ctx, c.SuggestTime, c.ContentType, cd.SubContentIDs(ctx))
	if err != nil {
		log.Warn(ctx, "check suggest time failed", log.Err(err), log.Any("req", c))
		return nil, err
	}

	err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{TeacherManualBatch: c.TeacherManualBatch})
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

	if c.SourceType == "" {
		c.SourceType = cm.getSourceType(ctx, c, cd)
	}

	//若为asset，直接发布
	//if the content is assets, publish immediately
	if c.ContentType == entity.ContentTypeAssets {
		publishStatus = entity.NewContentPublishStatus(entity.ContentStatusPublished)
		c.SelfStudy = false
		c.DrawActivity = false
		c.LessonType = ""
	}
	if c.ContentType == entity.ContentTypePlan {
		c.LessonType = ""
	}

	path := constant.FolderRootPath
	return &entity.Content{
		//ID:            utils.NewID(),
		ContentType:   c.ContentType,
		Name:          c.Name,
		Keywords:      strings.Join(c.Keywords, constant.StringArraySeparator),
		Description:   c.Description,
		Thumbnail:     c.Thumbnail,
		SuggestTime:   c.SuggestTime,
		Data:          c.Data,
		Extra:         c.Extra,
		LessonType:    c.LessonType,
		DirPath:       entity.NewPath(path),
		SelfStudy:     c.SelfStudy.Int(),
		DrawActivity:  c.DrawActivity.Int(),
		Outcomes:      strings.Join(c.Outcomes, constant.StringArraySeparator),
		Author:        operator.UserID,
		Creator:       operator.UserID,
		LockedBy:      constant.LockedByNoBody,
		Org:           operator.OrgID,
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
	if data.Description != "" {
		content.Description = data.Description
	}
	if data.Thumbnail != "" {
		content.Thumbnail = data.Thumbnail
	}
	if data.Outcomes != nil {
		content.Outcomes = strings.Join(data.Outcomes, constant.StringArraySeparator)
	}
	if data.Extra != "" {
		content.Extra = data.Extra
	}
	if len(data.Keywords) > 0 {
		content.Keywords = strings.Join(data.Keywords, constant.StringArraySeparator)
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

	//Asset修改后直接发布
	//if the content is assets, publish immediately after update
	if content.ContentType.IsAsset() {
		content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusPublished)
	}

	if data.SourceType != "" {
		content.SourceType = data.SourceType
	}

	//检查data
	if data.Data != "" {
		cd, err := cm.CreateContentData(ctx, data.ContentType, data.Data)
		if err != nil {
			return nil, ErrInvalidContentData
		}
		err = cd.Validate(ctx, content.ContentType)
		if err != nil {
			return nil, err
		}

		//TODO:For authed content => update contentdata sub content versions => done
		//update for version
		err = cd.PrepareVersion(ctx)
		if err != nil {
			log.Error(ctx, "can't update contentdata version for details", log.Err(err))
			return nil, ErrParseContentDataDetailsFailed
		}
		err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{TeacherManualBatch: data.TeacherManualBatch})
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
		err = cm.checkSuggestTime(ctx, data.SuggestTime, data.ContentType, cd.SubContentIDs(ctx))
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
	content.CopySourceID = content.ID
	content.Version = content.Version + 1
	content.ID = ""
	content.LockedBy = constant.LockedByNoBody
	content.Author = user.UserID
	//content.Author = user.UserID
	//content.Org = user.OrgID
	content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusDraft)
	now := time.Now()
	content.UpdateAt = now.Unix()
	content.CreateAt = now.Unix()

	return content
}

func (cm ContentModel) prepareCopyContentParams(ctx context.Context, content *entity.Content, user *entity.Operator) *entity.Content {
	content.Version = 1
	content.ID = ""
	content.LockedBy = constant.LockedByNoBody
	content.Author = user.UserID
	content.Org = user.OrgID
	content.CopySourceID = content.ID
	content.PublishStatus = entity.NewContentPublishStatus(entity.ContentStatusPublished)
	return content
}

func (cm ContentModel) prepareDeleteContentParams(ctx context.Context, content *entity.Content, publishStatus entity.ContentPublishStatus, user *entity.Operator) *entity.Content {
	//删除的时候不去掉路径信息
	//delete the dir path info
	// content.DirPath = constant.FolderRootPath

	//assets则隐藏
	//if content is assets, hide it
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

func (cm *ContentModel) getContentPath(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	dirPath := content.DirPath

	//if content.SourceID != "" {
	//	sourceContent, err := da.GetContentDA().GetContentByID(ctx, tx, content.SourceID)
	//	if err != nil {
	//		log.Error(ctx, "get source content failed", log.Err(err), log.Any("content", content))
	//		return err
	//	}
	//	dirPath = sourceContent.DirPath
	//	content.DirPath = dirPath
	//	return nil
	//}

	contentPath, err := GetFolderModel().MustGetPath(ctx, tx, entity.OwnerTypeOrganization, entity.FolderItemTypeFolder, dirPath.String(), entity.FolderPartitionMaterialAndPlans, user)
	if err != nil {
		log.Error(ctx, "search content folder failed",
			log.Err(err), log.Any("content", content))
		return err
	}
	//若路径不存在，则放到根目录
	//if dir path is not exists, put it into root path
	content.DirPath = entity.NewPath(contentPath)
	return nil
}

func (cm *ContentModel) preparePublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//若content为archive，则直接发布
	//if content is archived, publish it immediately
	if content.PublishStatus == entity.ContentStatusArchive {
		content.PublishStatus = entity.ContentStatusPublished
		content.UpdateAt = time.Now().Unix()
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
