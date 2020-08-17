package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"sync"
)

var (
	_contentVersionOnce sync.Once
	_contentVersionModel IContentVersionModel
)

var (
	ErrUnknownContentBaseVersion = errors.New("unknown content base version")
	ErrUnknownContentVersion = errors.New("unknown content version")
)


type IContentVersionModel interface {
	//AddNewContent
	AddContentVersion(ctx context.Context, tx *dbo.DBContext, baseCid, cid string) error

	//获取内容的最新版本
	GetContentLatestVersion(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentVersionData, error)

	//获取Content版本
	GetContentVersions(ctx context.Context, tx *dbo.DBContext, cid string) ([]*entity.ContentVersionData, error)

	//获取固定版本的Content
	GetContentByVersion(ctx context.Context, tx *dbo.DBContext, cid string, version int) (*entity.ContentVersionData, error)

	RemoveContentByContentId(ctx context.Context, tx *dbo.DBContext, cid string) error

	//检查该内容是否为最新版本
	IsContentLatest(ctx context.Context, tx *dbo.DBContext, cid string) (bool, error)
}

type ContentVersionModel struct {
	contentVersionDA da.IContentVersionDA
}

func (cv ContentVersionModel) AddContentVersion(ctx context.Context, tx *dbo.DBContext, baseCid, cid string) error {
	if baseCid != "" {
		//全新的版本
		_, err := cv.contentVersionDA.AddContentVersionRecord(ctx, tx, entity.ContentVersion{
			ContentId: cid,
			LastId:    "",
			MainId:    cid,
			SourceId: 	cid,
			Version:   1,
		})
		if err != nil{
			return err
		}
		return nil
	}

	//老版本
	//获取基础版版本信息
	_, res, err := cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		ContentIds: []string{baseCid},
	})
	if err != nil{
		return err
	}
	if len(res) < 1 {
		return ErrUnknownContentBaseVersion
	}
	//插入新版版本信息
	_, err = cv.contentVersionDA.AddContentVersionRecord(ctx, tx, entity.ContentVersion{
		ContentId: cid,
		LastId:    res[0].ContentId,
		MainId:    res[0].MainId,
		SourceId:	res[0].SourceId,
		Version:   res[0].Version + 1,
	})
	if err != nil{
		return err
	}

	return nil
}

func (cv ContentVersionModel) GetContentLatestVersion(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentVersionData, error) {
	versionList, err := cv.GetContentVersions(ctx, tx, cid)
	if err != nil {
		return nil ,err
	}
	return versionList[0], nil
}

func (cv ContentVersionModel) GetContentByVersion(ctx context.Context, tx *dbo.DBContext, cid string, version int) (*entity.ContentVersionData, error){
	res, err := cv.GetContentVersions(ctx, tx, cid)
	if err != nil{
		return nil, err
	}
	for i := range res {
		if res[i].Version == version {
			return res[i], nil
		}
	}
	return nil, ErrUnknownContentVersion
}
func (cv ContentVersionModel) RemoveContentByContentId(ctx context.Context, tx *dbo.DBContext, cid string) error {
	_, res, err := cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		ContentIds: []string{cid},
	})
	if err != nil{
		return err
	}
	//没有数据，则不删除
	if len(res) < 1{
		return nil
	}
	err = cv.contentVersionDA.RemoveContentVersionRecord(ctx, tx, res[0].Id)
	if err != nil{
		return err
	}
	return nil
}
func (cv *ContentVersionModel) IsContentLatest(ctx context.Context, tx *dbo.DBContext, cid string) (bool, error) {
	_, res, err := cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		ContentIds: []string{cid},
	})
	if err != nil{
		return false, err
	}
	if len(res) < 1 {
		return false, ErrUnknownContentBaseVersion
	}
	currentVersion := res[0].Version

	_, res, err = cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		MainIds:    []string{res[0].SourceId},
	})

	if err != nil{
		return false, err
	}
	if len(res) < 1 {
		return false, ErrUnknownContentVersion
	}
	for i := range res {
		if currentVersion < res[i].Version {
			return false, nil
		}
	}
	return true, nil
}

func (cv ContentVersionModel) GetContentVersions(ctx context.Context, tx *dbo.DBContext, cid string) ([]*entity.ContentVersionData, error) {
	_, res, err := cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		ContentIds: []string{cid},
	})
	if err != nil{
		return nil, err
	}
	if len(res) < 1 {
		return nil, ErrUnknownContentBaseVersion
	}

	_, res, err = cv.contentVersionDA.SearchContentVersion(ctx, tx, da.ContentVersionCondition{
		MainIds:    []string{res[0].SourceId},
	})

	if err != nil{
		return nil, err
	}
	if len(res) < 1 {
		return nil, ErrUnknownContentVersion
	}
	contentVersionList := make([]*entity.ContentVersionData ,len(res))
	for i := range res {
		contentVersionList[i] =&entity.ContentVersionData{
			RecordId:  res[i].Id,
			ContentId: res[i].ContentId,
			Version:   res[i].Version,
		}
	}

	return contentVersionList, nil
}

func GetContentVersionModel() IContentVersionModel {
	_contentVersionOnce.Do(func() {
		_contentVersionModel = &ContentVersionModel{
			contentVersionDA:          da.GetContentVersionDA(),
		}
	})
	return _contentVersionModel
}