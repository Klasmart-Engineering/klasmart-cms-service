package model

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	mutex "github.com/KL-Engineering/kidsloop-cms-service/mutex"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

var (
	ErrNoAuth = errors.New("no auth to operate")

	ErrNoContentData                     = errors.New("no content data")
	ErrInvalidResourceID                 = errors.New("invalid resource id")
	ErrInvalidContentData                = errors.New("invalid content data")
	ErrMarshalContentDataFailed          = errors.New("marshal content data failed")
	ErrInvalidPublishStatus              = errors.New("invalid publish status")
	ErrPlanHasArchivedMaterials          = errors.New("plan has archived materials")
	ErrInvalidLockedContentPublishStatus = errors.New("invalid locked content publish status")
	ErrInvalidContentStatusToPublish     = errors.New("content status is invalid to publish")
	ErrNoContent                         = errors.New("no content")
	//ErrContentAlreadyLocked              = errors.New("content is already locked")
	ErrDeleteLessonInSchedule        = errors.New("can't delete lesson in schedule")
	ErrGetUnpublishedContent         = errors.New("unpublished content")
	ErrGetUnauthorizedContent        = errors.New("unauthorized content")
	ErrCloneContentFailed            = errors.New("clone content failed")
	ErrParseContentDataFailed        = errors.New("parse content data failed")
	ErrParseContentDataDetailsFailed = errors.New("parse content data details failed")
	ErrUpdateContentFailed           = errors.New("update contentdata into data access failed")
	ErrReadContentFailed             = errors.New("read content failed")
	ErrDeleteContentFailed           = errors.New("delete contentdata into data access failed")
	ErrInvalidVisibleScope           = errors.New("invalid visible scope")

	ErrSuggestTimeTooSmall = errors.New("suggest time is less than sub contents")

	ErrInvalidMaterialType    = errors.New("invalid material type")
	ErrInvalidSourceOrContent = errors.New("invalid content data source or content")

	ErrEmptyTeacherManual = errors.New("empty teacher manual")

	ErrBadRequest         = errors.New("bad request")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrNoRejectReason     = errors.New("no reject reason")

	ErrInvalidSelectForm   = errors.New("invalid select form")
	ErrUserNotFound        = errors.New("user not found in locked by")
	ErrNoPermissionToQuery = errors.New("no permission to query")

	ErrMoveToDifferentPartition = errors.New("move to different parititon")
)

type ErrContentAlreadyLocked struct {
	LockedBy *external.User
}

func (e *ErrContentAlreadyLocked) Error() string {
	return "content is already locked"
}
func NewErrContentAlreadyLocked(ctx context.Context, lockedBy string, operator *entity.Operator) error {
	user, err := external.GetUserServiceProvider().Get(ctx, operator, lockedBy)
	if err != nil {
		log.Error(ctx, "external.GetUserServiceProvider().Get error",
			log.Err(err),
			log.String("lockedBy", lockedBy))
		return ErrUserNotFound
	}

	log.Debug(ctx, "locked by user", log.Any("user", user))
	return &ErrContentAlreadyLocked{LockedBy: user}
}

type SubContentsWithName struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Data       ContentData `json:"data"`
	OutcomeIDs []string    `json:"outcome_ids"`
}

type IContentModel interface {
	CreateContent(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (string, error)
	UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error
	PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error
	PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error
	LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)
	DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error
	CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error)

	PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error
	DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error

	GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)
	GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error)
	GetLatestContentIDByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error)
	GetPastContentIDByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	GetRawContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error)
	GetRawContentByIDListWithVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentWithVisibilitySettings, error)

	GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error)
	GetContentNameByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error)
	GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator, ignorePermissionFilter bool) ([]*SubContentsWithName, error)
	GetContentsSubContentsMapByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*SubContentsWithName, error)

	UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid string, reason []string, remark, status string) error
	CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error

	CountUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, error)
	SearchUserPrivateFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error)
	SearchUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error)
	SearchUserFolderContentSlim(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error)
	SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)

	SearchSimplifyContentInternal(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentInternalConditionRequest) (*entity.ContentSimplifiedList, error)

	GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	GetVisibleContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error)
	ContentDataCount(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentStatisticsInfo, error)
	GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error)

	IsContentsOperatorByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (bool, error)

	BatchUpdateContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, path entity.Path) error
	BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error

	//For authed content
	SearchSharedContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error)
	SearchSharedContentV2(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (response entity.QuerySharedContentV2Response, err error)

	CreateContentData(ctx context.Context, contentType entity.ContentType, data string) (ContentData, error)
	ConvertContentObj(ctx context.Context, tx *dbo.DBContext, obj *entity.Content, operator *entity.Operator) (*entity.ContentInfo, error)
	BatchConvertContentObj(ctx context.Context, tx *dbo.DBContext, objs []*entity.Content, operator *entity.Operator) ([]*entity.ContentInfo, error)

	ConvertContentObjWithProperties(ctx context.Context, obj *entity.Content, properties []*entity.ContentProperty) (*entity.ContentInfo, error)

	PublishContentWithAssetsTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error
	LockContentTx(ctx context.Context, cid string, user *entity.Operator) (string, error)
	PublishContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error
	PublishContentTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error
	DeleteContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error
	DeleteContentTx(ctx context.Context, cid string, user *entity.Operator) error

	GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest) (response *entity.GetLessonPlansCanScheduleResponse, err error)

	ContentsVisibleMap(ctx context.Context, cids []string, operator *entity.Operator) (map[string]entity.ContentAuth, error)
	GetSharedContents(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest) ([]*entity.Content, error)
	CheckContentVisible(ctx context.Context, cid string, operator *entity.Operator) (bool, error)
	UpdateSharedContentsCount(ctx context.Context, tx *dbo.DBContext, cids []string, operator *entity.Operator) error

	GetSpecifiedLessonPlan(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, planID string, materialIDs []string, withAP bool) (*entity.ContentInfoWithDetails, error)

	GetContentByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentInfoInternal, error)
	GetContentsSubContentsMapByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*entity.ContentInfoInternal, error)
	GetLatestContentIDMapByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string) (map[string]string, error)

	CleanCache(ctx context.Context)
}

func (cm *ContentModel) GetSpecifiedLessonPlan(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, planID string, materialIDs []string, withAP bool) (*entity.ContentInfoWithDetails, error) {
	latestIDs, err := cm.GetLatestContentIDByIDList(ctx, tx, []string{planID})
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan failed", log.Err(err),
			log.Any("operator", operator),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, err
	}
	if len(latestIDs) != 1 {
		log.Error(ctx, "GetSpecifiedLessonPlan failed", log.Err(err),
			log.Any("operator", operator),
			log.Strings("latest_ids", latestIDs),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, constant.ErrInternalServer
	}
	hasPermission, err := GetContentPermissionMySchoolModel().CheckGetContentPermission(ctx, latestIDs[0], operator)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan has no permission",
			log.Err(err),
			log.Any("operator", operator),
			log.Strings("latest_ids", latestIDs),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, err
	}
	if !hasPermission {
		log.Debug(ctx, "GetSpecifiedLessonPlan has no permission",
			log.Any("operator", operator),
			log.Strings("latest_ids", latestIDs),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, constant.ErrForbidden
	}

	var cids []string
	cids = append(cids, planID)
	cids = append(cids, materialIDs...)
	contents, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan failed",
			log.Err(err),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, err
	}
	var lessonPlan *entity.Content
	latestKeyContents := make(map[string]*entity.Content)
	for _, c := range contents {
		if c.LatestID != "" {
			latestKeyContents[c.LatestID] = c
		} else {
			latestKeyContents[c.ID] = c
		}

		if c.ID == planID {
			lessonPlan = c
		}
	}

	if lessonPlan == nil {
		log.Error(ctx, "GetSpecifiedLessonPlan failed",
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", materialIDs))
		return nil, constant.ErrRecordNotFound
	}

	lessonPlanData := new(LessonData)
	err = lessonPlanData.Unmarshal(ctx, lessonPlan.Data)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan unmarshal lesson plan data failed",
			log.Err(err),
			log.Any("lesson_plan", lessonPlan))
		return nil, constant.ErrRecordNotFound
	}

	planDataMaterialIDs := lessonPlanData.SubContentIDs(ctx)

	materialContents, err := da.GetContentDA().GetContentByIDList(ctx, tx, planDataMaterialIDs)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan get material failed",
			log.Err(err),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", planDataMaterialIDs))
		return nil, err
	}
	originSpecialMap := make(map[string]string)
	for _, m := range materialContents {
		latestID := m.LatestID
		if latestID == "" {
			latestID = m.ID
		}
		var special string
		if v, ok := latestKeyContents[latestID]; ok {
			special = v.ID
		} else {
			//special = m.ID
			log.Error(ctx, "GetSpecifiedLessonPlan map special failed",
				log.String("lesson_plan_id", planID),
				log.Strings("material_ids", planDataMaterialIDs),
				log.Any("db_material", m))
			return nil, constant.ErrInternalServer
		}
		originSpecialMap[m.ID] = special
	}

	lessonPlanData.LessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
		l.MaterialId = originSpecialMap[l.MaterialId]
	})
	lessonPlan.Data, err = lessonPlanData.Marshal(ctx)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan map ConvertContentObj failed",
			log.Err(err),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", planDataMaterialIDs),
			log.Any("lesson_plan_data", lessonPlanData),
			log.Any("lesson_plan", lessonPlan))
		return nil, err
	}

	content, err := cm.ConvertContentObj(ctx, tx, lessonPlan, operator)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan ConvertContentObj failed",
			log.Err(err),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", planDataMaterialIDs),
			log.Any("lesson_plan", lessonPlan))
		return nil, err
	}
	contentData, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan CreateContentData failed",
			log.Err(err),
			log.String("lesson_plan_id", planID),
			log.Strings("material_ids", planDataMaterialIDs),
			log.Any("lesson_plan", lessonPlan))
		return nil, err
	}

	err = contentData.PrepareResult(ctx, tx, content, operator, false)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan can't get content data for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	filledContentData, err := contentData.Marshal(ctx)
	if err != nil {
		log.Error(ctx, "GetSpecifiedLessonPlan can't marshal content data for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	content.Data = filledContentData

	if withAP {
		contentWithDetails, err := cm.buildContentWithDetails(ctx, []*entity.ContentInfo{content}, true, operator)
		if err != nil {
			log.Error(ctx, "GetSpecifiedLessonPlan can't parse content data", log.Err(err))
			return nil, ErrReadContentFailed
		}
		if len(contentWithDetails) < 1 {
			return &entity.ContentInfoWithDetails{
				ContentInfo: *content,
			}, nil
		}
		return contentWithDetails[0], nil
	}
	return &entity.ContentInfoWithDetails{ContentInfo: *content}, nil
}

type ContentModel struct {
	sharedContentQueryCache *utils.LazyRefreshCache
}

func (cm *ContentModel) UpdateSharedContentsCount(ctx context.Context, tx *dbo.DBContext, contentIDs []string, operator *entity.Operator) error {
	contents, err := da.GetContentDA().GetContentByIDList(ctx, tx, contentIDs)
	if err != nil {
		log.Error(ctx, "UpdateSharedContentsCount failed", log.Err(err), log.Strings("content_ids", contentIDs))
		return err
	}

	allParentIDs := make([]string, 0)
	allAncestorIDs := make([]string, 0)
	for i := range contents {
		if contents[i].ContentType == entity.ContentTypeMaterial || contents[i].ContentType == entity.ContentTypePlan {
			fids := contents[i].DirPath.Parents()

			if len(fids) > 0 && fids[len(fids)-1] != constant.FolderRootPath && fids[len(fids)-1] != "" {
				allParentIDs = append(allParentIDs, fids[len(fids)-1])
			}
		}
		if contents[i].DirPath.Parent() != constant.FolderRootPath && contents[i].DirPath.Parent() != "" {
			allAncestorIDs = append(allAncestorIDs, contents[i].DirPath.Parents()...)
		}
	}

	err = GetFolderModel().BatchUpdateFolderItemCount(ctx, tx, allParentIDs)
	if err != nil {
		log.Error(ctx, "UpdateSharedContentsCount BatchUpdateFolderItemCount failed",
			log.Err(err),
			log.Any("parents", allParentIDs),
			log.Any("contents", contents))
		return err
	}
	err = GetFolderModel().BatchUpdateAncestorEmptyField(ctx, tx, allAncestorIDs)
	if err != nil {
		log.Error(ctx, "UpdateSharedContentsCount BatchUpdateAncestorEmptyField failed",
			log.Err(err),
			log.Any("ancestor", allAncestorIDs),
			log.Any("contents", contents))
		return err
	}
	return nil
}

func (cm *ContentModel) CheckContentVisible(ctx context.Context, cid string, operator *entity.Operator) (bool, error) {
	visibilities, err := cm.GetSharedContents(ctx, operator, &entity.ContentConditionRequest{
		ContentIDs:  entity.NullStrings{Strings: []string{cid}, Valid: true},
		AuthedOrgID: entity.NullStrings{Strings: []string{operator.OrgID, constant.ShareToAll}},
	})
	if err != nil {
		log.Error(ctx, "No permission",
			log.Err(err),
			log.String("cid", cid),
			log.Any("op", operator))
		return false, err
	}
	if len(visibilities) > 0 {
		return true, nil
	}
	return false, nil
}

func (cm *ContentModel) ContentsVisibleMap(ctx context.Context, cids []string, operator *entity.Operator) (map[string]entity.ContentAuth, error) {
	tx := dbo.MustGetDB(ctx)
	contents, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "ContentsVisibleMap failed",
			log.Err(err),
			log.Any("op", operator),
			log.Strings("contents", cids))
		return nil, err
	}

	authMap := make(map[string]entity.ContentAuth)
	for _, v := range cids {
		authMap[v] = entity.ContentUnauthed
	}
	pendingAuthContentIDs := make([]string, 0)

	contentLatestIDMap := make(map[string]string)
	contentLatestIDRevert := make(map[string]string)
	for i := range contents {
		if contents[i].LatestID == "" {
			contentLatestIDMap[contents[i].ID] = contents[i].ID
			contentLatestIDRevert[contents[i].ID] = contents[i].ID
		} else {
			contentLatestIDMap[contents[i].ID] = contents[i].LatestID
			contentLatestIDRevert[contents[i].LatestID] = contents[i].ID
		}
	}

	for i := range contents {
		if contents[i].Org == operator.OrgID {
			authMap[contents[i].ID] = entity.ContentAuthed
		} else {
			authMap[contents[i].ID] = entity.ContentUnauthed
			//search auth contents use latest id
			pendingAuthContentIDs = append(pendingAuthContentIDs, contentLatestIDMap[contents[i].ID])
		}
	}

	if len(pendingAuthContentIDs) > 0 {
		paths, err := da.GetFolderDA().GetSharedContentParentPath(ctx, tx, []string{operator.OrgID, constant.ShareToAll})
		if err != nil {
			return nil, err
		}
		if len(paths) <= 0 {
			log.Debug(ctx, "ContentsVisibleMap no shared",
				log.Any("result", authMap))
			return authMap, nil
		}
		visibleContents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
			IDS:        entity.NullStrings{Strings: pendingAuthContentIDs, Valid: true},
			ParentPath: entity.NullStrings{Strings: paths, Valid: len(paths) > 0},
		})
		if err != nil {
			log.Error(ctx, "ContentsVisibleMap failed",
				log.Err(err),
				log.Any("op", operator),
				log.Strings("contents", cids),
				log.Strings("pending_auth_ids", pendingAuthContentIDs),
				log.Strings("parent_path", paths))
			return nil, err
		}

		for i := range visibleContents {
			//result revert to current id
			authMap[contentLatestIDRevert[visibleContents[i].ID]] = entity.ContentAuthed
		}
	}

	return authMap, nil
}

func (cm *ContentModel) GetSharedContents(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest) ([]*entity.Content, error) {
	tx := dbo.MustGetDB(ctx)
	paths, err := da.GetFolderDA().GetSharedContentParentPath(ctx, tx, cond.AuthedOrgID.Strings)
	if err != nil {
		return nil, err
	}
	if len(paths) <= 0 {
		log.Debug(ctx, "GetSharedContents: no shared", log.Any("op", op), log.Any("cond", cond))
		return []*entity.Content{}, nil
	}
	contents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IDS:        cond.ContentIDs,
		ParentPath: entity.NullStrings{Strings: paths, Valid: len(paths) > 0},
	})
	if err != nil {
		log.Error(ctx, "GetSharedContents failed",
			log.Any("condition", cond),
			log.Strings("parents_path", paths),
			log.Err(err))
		return nil, err
	}
	return contents, err
}

func (cm *ContentModel) handleSourceContent(ctx context.Context, tx *dbo.DBContext, contentID, sourceID string) error {
	sourceContent, err := da.GetContentDA().GetContentByID(ctx, tx, sourceID)
	if err != nil {
		return err
	}
	sourceContent.PublishStatus = entity.ContentStatusHidden
	sourceContent.LatestID = contentID
	//解锁source content
	//Unlock source content
	sourceContent.LockedBy = constant.LockedByNoBody
	err = da.GetContentDA().UpdateContent(ctx, tx, sourceID, *sourceContent)
	if err != nil {
		log.Error(ctx, "update source content failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	log.Debug(ctx, "update source content successfully", log.Any("content", sourceContent))

	//更新所有latestID为sourceContent的Content
	//Update all sourceContent latestID fields
	oldContents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		LatestID: sourceContent.ID,
	})
	if err != nil {
		log.Error(ctx, "update old content failed", log.Err(err), log.String("SourceID", sourceContent.ID))
		return ErrUpdateContentFailed
	}

	oldContentIDs := make([]string, len(oldContents))
	for i := range oldContents {
		oldContents[i].LockedBy = constant.LockedByNoBody
		oldContents[i].PublishStatus = entity.ContentStatusHidden
		oldContents[i].LatestID = contentID
		err = da.GetContentDA().UpdateContent(ctx, tx, oldContents[i].ID, *oldContents[i])
		if err != nil {
			log.Error(ctx, "update old content failed", log.Err(err), log.String("OldID", oldContents[i].ID))
			return ErrUpdateContentFailed
		}

		log.Debug(ctx, "update old content successfully", log.Any("content", oldContents[i]))

		oldContentIDs[i] = oldContents[i].ID
	}
	return nil
}

func (cm *ContentModel) doPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.preparePublishContent(ctx, tx, content, user)
	if err != nil {
		log.Error(ctx, "prepare publish failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	if content.DirPath.Parent() != constant.FolderRootPath && content.DirPath.Parent() != "" {
		_, err := GetFolderModel().GetFolderByIDTx(ctx, tx, content.DirPath.Parent(), user)
		if err != nil && err == ErrResourceNotFound {
			content.DirPath = constant.FolderRootPath
			content.ParentFolder = content.DirPath.Parent()
		}
		if err != nil && err != ErrResourceNotFound {
			log.Error(ctx, "doPublishContent: check parent failed", log.Err(err), log.Any("content", content), log.String("uid", user.UserID))
			return err
		}
	}

	err = da.GetContentDA().UpdateContent(ctx, tx, content.ID, *content)
	if err != nil {
		log.Error(ctx, "update lesson plan failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return ErrUpdateContentFailed
	}

	if content.PublishStatus == entity.ContentStatusPublished {
		//更新content的path
		err := cm.getContentPath(ctx, tx, content, user)
		if err != nil {
			log.Error(ctx, "getContentPath failed",
				log.Any("content", content),
				log.Any("user", user))
			return err
		}
		if content.DirPath.Parent() != constant.FolderRootPath && content.DirPath.Parent() != "" {
			err = GetFolderModel().BatchUpdateFolderItemCount(ctx, tx, []string{content.DirPath.Parent()})
			if err != nil {
				log.Error(ctx, "BatchUpdateFolderItemCount failed",
					log.Any("content", content),
					log.Any("user", user))
				return err
			}
			err = GetFolderModel().BatchUpdateAncestorEmptyField(ctx, tx, content.DirPath.Parents())
			if err != nil {
				log.Error(ctx, "doPublishContent: BatchUpdateAncestorEmptyField failed",
					log.Err(err),
					log.Any("content", content),
					log.String("uid", user.UserID))
				return err
			}
		}
	}

	//If scope changed, refresh visibility settings
	if scope != nil && len(scope) > 0 {
		err = cm.refreshContentVisibilitySettings(ctx, tx, content.ID, scope)
		if err != nil {
			log.Error(ctx, "refreshContentVisibilitySettings failed",
				log.Err(err),
				log.String("cid", content.ID),
				log.Strings("scope", scope),
				log.String("uid", user.UserID))
			return ErrUpdateContentFailed
		}
	}

	return nil
}

func (cm ContentModel) checkContentInfo(ctx context.Context, c entity.CreateContentRequest, op *entity.Operator) error {
	if c.LessonType != "" {
		_, err := GetLessonTypeModel().GetByID(ctx, c.LessonType)
		if err != nil {
			log.Error(ctx, "lesson type invalid", log.Any("data", c), log.Err(err))
			return ErrInvalidContentData
		}
	}
	err := c.Validate()
	if err != nil {
		log.Error(ctx, "asset no need to check", log.Any("data", c), log.Err(err))
		return err
	}
	err = c.ContentType.Validate()
	if err != nil {
		log.Error(ctx, "content type invalid", log.Any("data", c), log.Err(err))
		return err
	}
	if len(c.Category) == 0 ||
		c.Program == "" {
		log.Error(ctx, "select form invalid", log.Any("data", c), log.Err(err))
		return ErrInvalidSelectForm
	}

	_, err = prepareAllNeededName(ctx, op, entity.ExternalOptions{
		OrgIDs:     []string{op.OrgID},
		UsrIDs:     []string{op.UserID},
		ProgIDs:    []string{c.Program},
		SubjectIDs: c.Subject,
		CatIDs:     c.Category,
		SubcatIDs:  c.SubCategory,
		GradeIDs:   c.Grade,
		AgeIDs:     c.Age,
	})
	if err != nil {
		log.Error(ctx, "checkContentInfo: prepareAllNeededName failed", log.Err(err), log.Any("op", op), log.Any("req", c))
		return err
	}
	return nil
}

func (cm ContentModel) checkUpdateContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) (*entity.Content, error) {

	//若为asset，直接发布
	//if the content is assets, publish it immediate
	if content.ContentType.IsAsset() {
		log.Info(ctx, "asset no need to check", log.String("cid", content.ID))
		return content, nil
	}

	if content.PublishStatus == entity.ContentStatusPending ||
		content.PublishStatus == entity.ContentStatusArchive ||
		content.PublishStatus == entity.ContentStatusHidden ||
		content.PublishStatus == entity.ContentStatusAttachment ||
		content.PublishStatus == entity.ContentStatusPublished {
		return nil, ErrInvalidPublishStatus
	}
	return content, nil
}

func (cm ContentModel) checkPublishContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//若content为已发布状态或发布中状态，则创建新content
	//if content is published or pending, create a new content
	if content.PublishStatus != entity.ContentStatusDraft && content.PublishStatus != entity.ContentStatusRejected &&
		content.PublishStatus != entity.ContentStatusArchive {
		//报错
		//error
		log.Warn(ctx, "invalid content status", log.Any("content", content))
		return ErrInvalidContentStatusToPublish
	}

	contentData, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Any("contentData", contentData), log.Err(err))
		return err
	}
	subContentIDs := contentData.SubContentIDs(ctx)
	//No sub content, no need to check
	if len(subContentIDs) < 1 {
		return nil
	}

	contentList, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IDS: entity.NullStrings{
			Strings: subContentIDs,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "search content data failed", log.Any("IDS", subContentIDs), log.Err(err))
		return err
	}
	err = cm.checkPublishContentChildren(ctx, content, contentList)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentModel) convertAuthedOrgIDToParentsPath(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest) error {
	if !condition.AuthedOrgID.Valid || len(condition.AuthedOrgID.Strings) <= 0 {
		return nil
	}
	parents, err := da.GetFolderDA().GetSharedContentParentPath(ctx, tx, condition.AuthedOrgID.Strings)
	if err != nil {
		return err
	}
	condition.ParentsPath.Strings = parents
	condition.ParentsPath.Valid = true
	return nil
}

func (cm *ContentModel) searchContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	log.Debug(ctx, "search content ", log.Any("condition", condition), log.String("uid", user.UserID))
	err := cm.convertAuthedOrgIDToParentsPath(ctx, tx, condition)
	if err != nil {
		return 0, nil, err
	}
	count, objs, err := da.GetContentDA().SearchContent(ctx, tx, cm.conditionRequestToCondition(*condition))
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	response, err := cm.BatchConvertContentObjForSearchContent(ctx, tx, objs, user)
	if err != nil {
		log.Error(ctx, "Can't parse contentdata, contentID: %v, error: %v", log.Any("objs", objs), log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	contentWithDetails, err := cm.buildContentWithDetailsForSearchContent(ctx, response, user)
	if err != nil {
		log.Error(ctx, "build content details failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	return count, contentWithDetails, nil
}
func (cm *ContentModel) searchContentUnsafe(ctx context.Context, tx *dbo.DBContext, condition dbo.Conditions, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	count, objs, err := da.GetContentDA().SearchContentUnSafe(ctx, tx, condition)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	response, err := cm.BatchConvertContentObj(ctx, tx, objs, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	contentWithDetails, err := cm.buildContentWithDetails(ctx, response, false, user)
	if err != nil {
		log.Error(ctx, "build content details failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}

	return count, contentWithDetails, nil
}

func (cm *ContentModel) CheckCreateContentParams(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (*entity.Content, error) {
	//检查数据信息是否正确
	//valid the data
	c.Trim()
	if c.ContentType.IsAsset() {
		// use operator's org id as asset publish scope, maybe not right...
		c.PublishScope = []string{operator.OrgID}
	}
	err := cm.checkContentInfo(ctx, c, operator)
	if err != nil {
		log.Warn(ctx, "check content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, err
	}
	//组装要创建的内容
	//construct the new content structure
	content, err := cm.prepareCreateContentParams(ctx, c, operator)
	if err != nil {
		log.Warn(ctx, "prepare content failed", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
		return nil, err
	}
	return content, nil
}

func (cm *ContentModel) CreateContent(ctx context.Context, c entity.CreateContentRequest, operator *entity.Operator) (string, error) {
	content, err := cm.CheckCreateContentParams(ctx, c, operator)
	if err != nil {
		return "", err
	}
	log.Info(ctx, "create content")
	cid, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		//添加内容
		//do insert content into database
		now := time.Now()
		content.UpdateAt = now.Unix()
		content.CreateAt = now.Unix()
		pid, err := da.GetContentDA().CreateContent(ctx, tx, *content)
		if err != nil {
			log.Error(ctx, "can't create contentdata", log.Err(err), log.String("uid", operator.UserID), log.Any("data", c))
			return "", err
		}

		//Insert content properties
		err = cm.doCreateContentProperties(ctx, tx, entity.ContentProperties{
			ContentID:   pid,
			Program:     c.Program,
			Subject:     c.Subject,
			Category:    c.Category,
			SubCategory: c.SubCategory,
			Age:         c.Age,
			Grade:       c.Grade,
		}, false)
		if err != nil {
			log.Error(ctx, "doCreateContentProperties failed",
				log.Err(err),
				log.String("uid", operator.UserID),
				log.Any("data", c))
			return "", err
		}

		//Insert into visibility settings
		err = cm.insertContentVisibilitySettings(ctx, tx, pid, c.PublishScope)
		if err != nil {
			log.Error(ctx, "insertContentVisibilitySettings failed",
				log.Err(err),
				log.String("uid", operator.UserID),
				log.String("pid", pid),
				log.Any("data", c))
			return "", err
		}

		if content.ContentType.IsAsset() &&
			content.PublishStatus == entity.NewContentPublishStatus(entity.ContentStatusPublished) &&
			content.DirPath.Parent() != constant.FolderRootPath &&
			content.DirPath.Parent() != "" {
			err = GetFolderModel().BatchUpdateFolderItemCount(ctx, tx, []string{content.DirPath.Parent()})
			if err != nil {
				log.Error(ctx, "CreateContent: BatchUpdateFolderItemCount failed",
					log.Err(err),
					log.String("uid", operator.UserID),
					log.String("pid", pid),
					log.Any("content", content))
				return "", err
			}
		}
		if content.PublishStatus == entity.NewContentPublishStatus(entity.ContentStatusPublished) &&
			content.DirPath.Parent() != constant.FolderPathSeparator && content.DirPath.Parent() != "" {
			err = GetFolderModel().BatchUpdateAncestorEmptyField(ctx, tx, content.DirPath.Parents())
			if err != nil {
				log.Error(ctx, "CreateContent: BatchUpdateAncestorEmptyField failed",
					log.Err(err),
					log.String("uid", operator.UserID),
					log.String("pid", pid),
					log.Any("content", content))
				return "", err
			}
		}

		return pid, nil
	})
	if cid == nil {
		return "", err
	}
	return cid.(string), err
}

func (cm *ContentModel) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, data entity.CreateContentRequest, user *entity.Operator) error {
	data.Trim()
	if data.ContentType.IsAsset() {
		//Assets can't be updated
		return ErrInvalidContentType
	}

	err := cm.checkContentInfo(ctx, data, user)
	if err != nil {
		return err
	}
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}

	content, err = cm.checkUpdateContent(ctx, tx, content, user)
	if err != nil {
		log.Error(ctx, "check update content failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}
	//Check status
	//Check Authorization
	obj, err := cm.prepareUpdateContentParams(ctx, content, &data)
	if err != nil {
		log.Error(ctx, "can't prepare params contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}

	//更新数据库
	//do update data into database
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *obj)
	if err != nil {
		log.Error(ctx, "update contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID), log.Any("data", data))
		return err
	}
	//Insert content properties
	err = cm.doCreateContentProperties(ctx, tx, entity.ContentProperties{
		ContentID:   cid,
		Program:     data.Program,
		Subject:     data.Subject,
		Category:    data.Category,
		SubCategory: data.SubCategory,
		Age:         data.Age,
		Grade:       data.Grade,
	}, true)
	if err != nil {
		log.Error(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.String("uid", user.UserID),
			log.Any("data", data))
		return err
	}

	//若已发布，不能修改publishScope
	//If it is already published, can't update publish scope
	if content.PublishStatus == entity.ContentStatusDraft ||
		content.PublishStatus == entity.ContentStatusRejected {
		err = cm.refreshContentVisibilitySettings(ctx, tx, content.ID, data.PublishScope)
		if err != nil {
			log.Error(ctx, "refreshContentVisibilitySettings failed",
				log.Err(err),
				log.String("cid", cid),
				log.String("uid", user.UserID),
				log.Any("data", data))
			return err
		}
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid})
	return nil
}

func (cm *ContentModel) BatchUpdateContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, path entity.Path) error {
	err := da.GetContentDA().BatchUpdateContentPath(ctx, tx, cids, path)
	if err != nil {
		log.Error(ctx, "update content path failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm *ContentModel) BatchReplaceContentPath(ctx context.Context, tx *dbo.DBContext, cids []string, oldPath, path string) error {
	err := da.GetContentDA().BatchReplaceContentPath(ctx, tx, cids, oldPath, path)
	if err != nil {
		log.Error(ctx, "update content path failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	return nil
}

func (cm *ContentModel) UpdateContentPublishStatus(ctx context.Context, tx *dbo.DBContext, cid string, reason []string, remark, status string) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	content.PublishStatus = entity.NewContentPublishStatus(status)
	if status == entity.ContentStatusRejected && len(reason) < 1 && remark == "" {
		return ErrNoRejectReason
	}
	operator := &entity.Operator{
		UserID: content.Author,
		OrgID:  content.Org,
	}

	if status == entity.ContentStatusPublished {
		//更新content的path
		err = cm.getContentPath(ctx, tx, content, operator)
		if err != nil {
			return err
		}
	}

	rejectReason := strings.Join(reason, constant.StringArraySeparator)
	content.RejectReason = rejectReason
	content.Remark = remark
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		log.Error(ctx, "update contentdata scope failed", log.Err(err))
		return ErrUpdateContentFailed
	}

	if status == entity.ContentStatusPublished {
		if content.SourceID != "" {
			//处理source content
			//handle with source content
			err = cm.handleSourceContent(ctx, tx, content.ID, content.SourceID)
			if err != nil {
				return err
			}
		}
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	cm.CleanCache(ctx)

	return nil
}

func (cm *ContentModel) UnlockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return err
	}
	content.LockedBy = constant.LockedByNoBody
	return da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
}

func (cm *ContentModel) LockContentTx(ctx context.Context, cid string, user *entity.Operator) (string, error) {
	ncid, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		ncid, err := cm.LockContent(ctx, tx, cid, user)
		if err != nil {
			return nil, err
		}
		return ncid, nil
	})
	if ncid == nil {
		return "", err
	}
	return ncid.(string), err
}

func (cm *ContentModel) LockContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixContentLock, cid)
	if err != nil {
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err))
		return "", err
	}
	if content.ContentType.IsAsset() {
		log.Info(ctx, "asset handle", log.String("cid", cid))
		return "", ErrInvalidContentType
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Info(ctx, "invalid publish status", log.String("cid", cid))
		return "", ErrInvalidPublishStatus
	}

	//被自己锁住，则返回锁定id
	//if it is locked by current user, return cloned content id
	if content.LockedBy == user.UserID {
		data, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
			SourceID: cid,
		})
		if err != nil {
			log.Error(ctx, "search source content failed", log.String("cid", cid))
			return "", err
		}
		if len(data) < 1 {
			//被自己锁定且找不到content
			//if locked by current user, but content is not found, panic
			log.Info(ctx, "no content in source content", log.String("cid", cid))
			return "", ErrNoContent
		}
		if data[0].PublishStatus != entity.ContentStatusRejected && data[0].PublishStatus != entity.ContentStatusDraft {
			log.Info(ctx, "invalid locked content status", log.String("lock cid", data[0].ID), log.String("status", string(data[0].PublishStatus)), log.String("cid", cid))
			return "", ErrInvalidLockedContentPublishStatus
		}
		//找到data
		//find the data
		return data[0].ID, nil
	}

	//更新锁定状态
	//update lock status
	if content.LockedBy != "" && content.LockedBy != constant.LockedByNoBody {
		return "", NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}
	content.LockedBy = user.UserID
	//content.Author = user.UserID
	err = da.GetContentDA().UpdateContent(ctx, tx, cid, *content)
	if err != nil {
		return "", err
	}
	//克隆Content
	//clone content
	ccid, err := cm.CloneContent(ctx, tx, content.ID, user)
	if err != nil {
		return "", err
	}
	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.ID})

	//Update with new content
	return ccid, nil

}
func (cm *ContentModel) PublishContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.PublishContentBulk(ctx, tx, ids, user)
		if err != nil {
			return err
		}
		return nil
	})
}
func (cm *ContentModel) PublishContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	updateIDs := make([]string, 0)
	if len(ids) < 1 {
		return nil
	}
	contents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IDS: entity.NullStrings{
			Strings: ids,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "can't read content on delete contentdata", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
		return err
	}
	for i := range contents {
		if contents[i].ContentType.IsAsset() {
			log.Warn(ctx, "try to publish asset", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
			continue
		}
		err = cm.doPublishContent(ctx, tx, contents[i], nil, user)
		if err != nil {
			log.Error(ctx, "can't publish content", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
			return err
		}
		updateIDs = append(updateIDs, contents[i].ID)
		if contents[i].SourceID != "" {
			updateIDs = append(updateIDs, contents[i].SourceID)
		}
	}

	da.GetContentRedis().CleanContentCache(ctx, updateIDs)
	cm.CleanCache(ctx)
	return err
}

type searchSharedContentRequest struct {
	Condition *entity.ContentConditionRequest
	Operator  *entity.Operator
}

type searchSharedContentResponse struct {
	Total    int
	Contents []*entity.ContentInfoWithDetails
}

func (cm *ContentModel) SearchSharedContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	//set condition with authed flag
	condition.AuthedOrgID = entity.NullStrings{
		Strings: []string{user.OrgID, constant.ShareToAll},
		Valid:   true,
	}

	condition.PublishStatus = []string{entity.ContentStatusPublished}

	if !config.Get().RedisConfig.OpenCache {
		return cm.searchContent(ctx, tx, condition, user)
	}

	request := &searchSharedContentRequest{
		Condition: condition,
		Operator:  user,
	}

	response := &searchSharedContentResponse{Contents: []*entity.ContentInfoWithDetails{}}

	err := cm.sharedContentQueryCache.Get(ctx, request, response)
	if err != nil {
		return 0, nil, err
	}

	return response.Total, response.Contents, nil
}
func (cm *ContentModel) SearchSharedContentV2(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, op *entity.Operator) (response entity.QuerySharedContentV2Response, err error) {
	response, err = da.GetContentDA().SearchSharedContentV2(ctx, tx, condition, op)
	if err != nil {
		return
	}

	// fill author name
	var userIDs []string
	for _, item := range response.Items {
		userIDs = append(userIDs, item.Author)
	}
	if len(userIDs) <= 0 {
		return
	}
	users, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, userIDs)
	if err != nil {
		return
	}
	for _, item := range response.Items {
		if name, ok := users[item.Author]; ok {
			item.AuthorName = name
		}
	}
	return
}

func (cm *ContentModel) searchSharedContent(ctx context.Context, condition interface{}) (interface{}, error) {
	request, ok := condition.(*searchSharedContentRequest)
	if !ok {
		log.Error(ctx, "invalid request", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	total, records, err := cm.searchContent(ctx, dbo.MustGetDB(ctx), request.Condition, request.Operator)
	if err != nil {
		return nil, err
	}

	return &searchSharedContentResponse{Total: total, Contents: records}, nil
}

func (cm *ContentModel) PublishContentTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.PublishContent(ctx, tx, cid, scope, user)
		return err
	})
}
func (cm *ContentModel) PublishContent(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata for publishing", log.Err(err), log.String("cid", cid), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	if content.ContentType.IsAsset() {
		return ErrInvalidContentType
	}

	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	cm.CleanCache(ctx)

	return nil
}

func (cm *ContentModel) validatePublishContentWithAssets(ctx context.Context, content *entity.Content, user *entity.Operator) error {
	if content.ContentType != entity.ContentTypeMaterial && content.ContentType != entity.ContentTypePlan {
		return ErrInvalidContentType
	}
	//查看data
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	switch data := cd.(type) {
	case *MaterialData:
		if data.InputSource != entity.MaterialInputSourceDisk {
			log.Warn(ctx, "invalid material type", log.Err(err), log.String("uid", user.UserID), log.Any("data", data))
			return ErrInvalidMaterialType
		}
	case *LessonData:
		if len(data.TeacherManualBatch) < 1 {
			log.Warn(ctx, "empty teacher manual batch", log.Err(err), log.String("uid", user.UserID), log.Any("data", data))
			return ErrEmptyTeacherManual
		}

	default:
		log.Warn(ctx, "validate content data with content type 2 failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	return nil
}

func (cm *ContentModel) prepareForPublishMaterialsAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	//创建data对象
	//create content data object
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	//解析data的fileType
	err = cd.PrepareSave(ctx, entity.ExtraDataInRequest{})
	materialData, ok := cd.(*MaterialData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", content.ID))
		return ErrInvalidContentData
	}

	//创建assets data对象，并解析
	//create assets data object, and parse it
	assetsData := new(AssetsData)
	assetsData.Source = materialData.Source
	assetsDataJSON, err := assetsData.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrMarshalContentDataFailed
	}
	//创建assets
	//Create assets
	req := entity.CreateContentRequest{
		ContentType: entity.ContentTypeAssets,
		Name:        content.Name,
		Keywords:    utils.StringToStringArray(ctx, content.Keywords),
		Description: content.Description,
		Thumbnail:   content.Thumbnail,
		SuggestTime: content.SuggestTime,
		Data:        assetsDataJSON,

		Program:      contentProperties.Program,
		Subject:      contentProperties.Subject,
		Category:     contentProperties.Category,
		SubCategory:  contentProperties.SubCategory,
		Age:          contentProperties.Age,
		Grade:        contentProperties.Grade,
		PublishScope: []string{user.OrgID},
	}
	_, err = cm.CreateContent(ctx, req, user)
	if err != nil {
		log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
		return err
	}

	//更新content状态
	//update content status
	materialData.InputSource = entity.MaterialInputSourceDisk
	d, err := materialData.Marshal(ctx)
	if err != nil {
		log.Warn(ctx, "marshal content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("content", content))
		return err
	}

	content.Data = d
	return nil
}

func (cm *ContentModel) PublishContentWithAssetsTx(ctx context.Context, cid string, scope []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return cm.PublishContentWithAssets(ctx, tx, cid, scope, user)
	})
}
func (cm *ContentModel) PublishContentWithAssets(ctx context.Context, tx *dbo.DBContext, cid string, scope []string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content data for publishing", log.Err(err), log.String("cid", cid), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	switch content.ContentType {
	case entity.ContentTypeMaterial:
		err = cm.publishMaterialWithAssets(ctx, tx, content, scope, user)
		if err != nil {
			return err
		}
	case entity.ContentTypePlan:
		err = cm.doPublishPlanWithAssets(ctx, tx, content, scope, user)
		if err != nil {
			return err
		}
	default:
		log.Warn(ctx, "content invalid",
			log.String("cid", cid),
			log.Any("content", content))
		return ErrInvalidContentType
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{content.ID, content.SourceID})
	cm.CleanCache(ctx)

	return nil
}
func (cm *ContentModel) DeleteContentBulkTx(ctx context.Context, ids []string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.DeleteContentBulk(ctx, tx, ids, user)
		if err != nil {
			return err
		}
		return nil
	})
}
func (cm *ContentModel) DeleteContentBulk(ctx context.Context, tx *dbo.DBContext, ids []string, user *entity.Operator) error {
	if len(ids) < 1 {
		return nil
	}
	deletedIDs := make([]string, 0)
	deletedIDs = append(deletedIDs, ids...)
	contents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IDS: entity.NullStrings{
			Strings: ids,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "can't read content on delete contentdata", log.Err(err), log.Strings("ids", ids), log.String("uid", user.UserID))
		return err
	}

	parentIDs := make([]string, 0)
	ancestorIDs := make([]string, 0)
	for i := range contents {
		err = cm.doDeleteContent(ctx, tx, contents[i], user)
		if err != nil {
			return err
		}
		//record pending delete content id
		if contents[i].SourceID != "" {
			deletedIDs = append(deletedIDs, contents[i].SourceID)
		}

		if contents[i].DirPath.Parent() != constant.FolderRootPath && contents[i].DirPath.Parent() != "" {
			parentIDs = append(parentIDs, contents[i].DirPath.Parent())
			ancestorIDs = append(ancestorIDs, contents[i].DirPath.Parents()...)
		}
	}
	err = GetFolderModel().BatchUpdateFolderItemCount(ctx, tx, parentIDs)
	if err != nil {
		log.Error(ctx, "DeleteContentBulk: BatchUpdateFolderItemCount",
			log.Err(err),
			log.Strings("parentIDs", parentIDs),
			log.Strings("ids", ids),
			log.String("uid", user.UserID))
		return err
	}

	err = GetFolderModel().BatchUpdateAncestorEmptyField(ctx, tx, ancestorIDs)
	if err != nil {
		log.Error(ctx, "DeleteContentBulk: BatchUpdateAncestorEmptyField",
			log.Err(err),
			log.Strings("ancestors", ancestorIDs),
			log.Strings("ids", ids),
			log.String("uid", user.UserID))
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, deletedIDs)
	cm.CleanCache(ctx)

	return nil
}

func (cm *ContentModel) publishMaterialWithAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", content.ID), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}

	//准备发布（1.创建assets，2.修改contentdata, 3.发布assets）
	//preparing to publish (1.create assets 2.update content data)
	err = cm.prepareForPublishMaterialsAssets(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	//发布
	//do publish
	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentModel) doPublishPlanWithAssets(ctx context.Context, tx *dbo.DBContext, content *entity.Content, scope []string, user *entity.Operator) error {
	err := cm.validatePublishContentWithAssets(ctx, content, user)
	if err != nil {
		log.Error(ctx, "validate for publishing failed", log.Err(err), log.String("cid", content.ID), log.Strings("scope", scope), log.String("uid", user.UserID))
		return err
	}
	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, content.ID)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", content.ID))
		return ErrInvalidContentData
	}

	//create content data object
	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Warn(ctx, "create content data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}

	//parse data for lesson plan
	lessonData, ok := cd.(*LessonData)
	if !ok {
		log.Warn(ctx, "asset content data type failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
		return ErrInvalidContentData
	}
	log.Info(ctx,
		"publish plans with teacher manuals",
		log.Any("content", content),
		log.Any("data", lessonData.TeacherManualBatch))
	//create assets data object, and parse it
	for i := range lessonData.TeacherManualBatch {
		assetsData := new(AssetsData)
		assetsData.Source = SourceID(lessonData.TeacherManualBatch[i].ID)
		assetsDataJSON, err := assetsData.Marshal(ctx)
		if err != nil {
			log.Warn(ctx, "marshal assets data failed", log.Err(err), log.String("uid", user.UserID), log.Any("data", content))
			return ErrMarshalContentDataFailed
		}
		//创建assets
		req := entity.CreateContentRequest{
			ContentType:  entity.ContentTypeAssets,
			Name:         content.Name,
			Keywords:     append(utils.StringToStringArray(ctx, content.Keywords), constant.TeacherManualAssetsKeyword),
			Description:  content.Description,
			Thumbnail:    "",
			SuggestTime:  0,
			Program:      contentProperties.Program,
			Subject:      contentProperties.Subject,
			Category:     contentProperties.Category,
			SubCategory:  contentProperties.SubCategory,
			Age:          contentProperties.Age,
			Grade:        contentProperties.Grade,
			PublishScope: []string{user.OrgID},
			Data:         assetsDataJSON,
		}
		_, err = cm.CreateContent(ctx, req, user)
		if err != nil {
			log.Warn(ctx, "create assets failed", log.Err(err), log.String("uid", user.UserID), log.Any("req", req))
			return err
		}
	}
	//do publish
	err = cm.doPublishContent(ctx, tx, content, scope, user)
	if err != nil {
		return err
	}
	return nil
}

func (cm *ContentModel) checkDeleteContent(ctx context.Context, content *entity.Content) error {
	if content.PublishStatus == entity.ContentStatusArchive && content.ContentType == entity.ContentTypePlan {
		exist, err := GetScheduleModel().ExistScheduleByLessonPlanID(ctx, content.ID)
		if err != nil {
			return err
		}
		if exist {
			return ErrDeleteLessonInSchedule
		}
	}
	return nil
}

func (cm *ContentModel) doDeleteContent(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	if content.LockedBy != constant.LockedByNoBody && content.LockedBy != user.UserID {
		return NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}
	if content.PublishStatus == entity.ContentStatusPublished && content.LockedBy != constant.LockedByNoBody {
		return NewErrContentAlreadyLocked(ctx, content.LockedBy, user)
	}

	obj := cm.prepareDeleteContentParams(ctx, content, content.PublishStatus, user)

	err := da.GetContentDA().UpdateContent(ctx, tx, content.ID, *obj)
	if err != nil {
		log.Error(ctx, "delete contentdata failed", log.Err(err), log.String("cid", content.ID), log.String("uid", user.UserID))
		return err
	}

	log.Debug(ctx, "delete_content_not_update count",
		log.String("parent", content.DirPath.Parent()),
		log.Any("content", content))

	if content.DirPath.Parent() != constant.FolderRootPath && content.DirPath.Parent() != "" {
		err = GetFolderModel().BatchUpdateFolderItemCount(ctx, tx, []string{content.DirPath.Parent()})
		if err != nil {
			log.Error(ctx, "doDeleteContent: BatchUpdateFolderItemCount failed",
				log.Err(err),
				log.Any("content", content),
				log.String("uid", user.UserID))
			return err
		}
		if content.PublishStatus == entity.ContentStatusPublished {
			err = GetFolderModel().BatchUpdateAncestorEmptyField(ctx, tx, content.DirPath.Parents())
			if err != nil {
				log.Error(ctx, "doDeleteContent: BatchUpdateAncestorEmptyField failed",
					log.Err(err),
					log.Any("content", content),
					log.String("uid", user.UserID))
				return err
			}
		}
	}

	//解锁source content
	if content.SourceID != "" {
		err = cm.UnlockContent(ctx, tx, content.SourceID, user)
		if err != nil {
			log.Error(ctx, "unlock contentdata failed", log.Err(err), log.String("cid", content.SourceID), log.String("uid", user.UserID))
			return err
		}
	}
	return nil
}
func (cm *ContentModel) DeleteContentTx(ctx context.Context, cid string, user *entity.Operator) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := cm.DeleteContent(ctx, tx, cid, user)
		if err != nil {
			return err
		}
		return nil
	})
}
func (cm *ContentModel) DeleteContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) error {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "content not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read content on delete content", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return err
	}

	err = cm.doDeleteContent(ctx, tx, content, user)
	if err != nil {
		return err
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{cid, content.SourceID})
	cm.CleanCache(ctx)

	return nil
}

func (cm *ContentModel) CloneContent(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata on update contentdata", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", err
	}
	if content.ContentType.IsAsset() {
		return "", ErrInvalidContentType
	}

	//检查是否有克隆权限
	//check permission
	err = cm.CheckContentAuthorization(ctx, tx, &entity.Content{
		ID:            content.ID,
		PublishStatus: content.PublishStatus,
		Author:        content.Author,
		Org:           content.Org,
	}, user)
	if err != nil {
		log.Error(ctx, "no auth to read content for cloning", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", ErrCloneContentFailed
	}

	obj := cm.prepareCloneContentParams(ctx, content, user)

	id, err := da.GetContentDA().CreateContent(ctx, tx, *obj)
	if err != nil {
		log.Error(ctx, "clone contentdata failed", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return "", err
	}
	//load content properties
	contentProperties, err := cm.getContentProperties(ctx, cid)
	if err != nil {
		log.Warn(ctx, "getContentProperties failed",
			log.Err(err), log.String("cid", cid))
		return "", ErrInvalidContentData
	}
	log.Info(ctx, "getContentProperties",
		log.Any("contentProperties", contentProperties),
		log.String("cid", cid))

	contentProperties.ContentID = id
	err = cm.doCreateContentProperties(ctx, tx, *contentProperties, false)
	if err != nil {
		log.Warn(ctx, "doCreateContentProperties failed",
			log.Err(err),
			log.Any("contentProperties", contentProperties),
			log.String("cid", cid))
		return "", err
	}

	contentVisibilitySettings, err := cm.getContentVisibilitySettings(ctx, cid)
	if err != nil {
		log.Warn(ctx, "getContentVisibilitySettings failed",
			log.Err(err), log.String("cid", cid))
		return "", ErrInvalidContentData
	}

	log.Info(ctx, "getContentVisibilitySettings",
		log.Any("contentVisibilitySettings", contentVisibilitySettings),
		log.String("cid", cid))

	err = cm.insertContentVisibilitySettings(ctx, tx, id, contentVisibilitySettings.VisibilitySettings)
	if err != nil {
		log.Warn(ctx, "insertContentVisibilitySettings failed",
			log.Err(err), log.String("cid", cid),
			log.Strings("VisibilitySettings", contentVisibilitySettings.VisibilitySettings))
		return "", ErrInvalidContentData
	}

	da.GetContentRedis().CleanContentCache(ctx, []string{id, obj.ID})
	cm.CleanCache(ctx)

	return id, nil
}

func (cm *ContentModel) CheckContentAuthorization(ctx context.Context, tx *dbo.DBContext, content *entity.Content, user *entity.Operator) error {
	//if user.UserID == content.Author {
	//	return nil
	//}

	if content.PublishStatus == entity.ContentStatusAttachment ||
		content.PublishStatus == entity.ContentStatusHidden {
		return nil
	}

	if content.PublishStatus != entity.ContentStatusPublished {
		log.Error(ctx, "read unpublished content, userID: %v, contentID: %v", log.String("userID", user.UserID), log.String("contentID", content.ID))
		return ErrGetUnpublishedContent
	}
	return nil
}

func (cm *ContentModel) GetContentsSubContentsMapByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*SubContentsWithName, error) {
	objs, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	contentInfoMap := make(map[string][]*SubContentsWithName)
	for _, obj := range objs {
		cd, err := cm.CreateContentData(ctx, obj.ContentType, obj.Data)
		if err != nil {
			log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
			return nil, err
		}

		switch v := cd.(type) {
		case *LessonData:
			//存在子内容，则返回子内容
			//if the content contains sub contents, return sub contents
			content, err := cm.ConvertContentObj(ctx, tx, obj, user)
			if err != nil {
				log.Error(ctx, "can't parse contentdata", log.Err(err))
				return nil, ErrParseContentDataFailed
			}
			err = v.PrepareVersion(ctx)
			if err != nil {
				log.Error(ctx, "can't prepare version for sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			err = v.PrepareResult(ctx, tx, content, user, false)
			if err != nil {
				log.Error(ctx, "can't get sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			ret := make([]*SubContentsWithName, 0)
			v.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
				if l.Material != nil {
					cd0, err := cm.CreateContentData(ctx, l.Material.ContentType, l.Material.Data)
					if err != nil {
						log.Error(ctx, "can't parse sub content data",
							log.Err(err),
							log.Any("lesson", l),
							log.Any("subContent", l.Material))
						return
					}
					ret = append(ret, &SubContentsWithName{
						ID:         l.Material.ID,
						Name:       l.Material.Name,
						Data:       cd0,
						OutcomeIDs: l.Material.Outcomes,
					})
				}
			})
			contentInfoMap[obj.ID] = ret
		case *MaterialData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*SubContentsWithName{
				{
					ID:         obj.ID,
					Name:       obj.Name,
					Data:       v,
					OutcomeIDs: cm.parseContentOutcomes(ctx, obj),
				},
			}
			contentInfoMap[obj.ID] = ret
		case *AssetsData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*SubContentsWithName{
				{
					ID:         obj.ID,
					Name:       obj.Name,
					Data:       v,
					OutcomeIDs: cm.parseContentOutcomes(ctx, obj),
				},
			}
			contentInfoMap[obj.ID] = ret
		}
	}

	return contentInfoMap, nil
}

func (cm *ContentModel) GetContentSubContentsByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator, ignorePermissionFilter bool) ([]*SubContentsWithName, error) {
	obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	//获取最新数据
	//fetch newest data
	if obj.LatestID != "" && obj.LatestID != obj.ID {
		obj, err = da.GetContentDA().GetContentByID(ctx, tx, obj.LatestID)
		if err != nil {
			log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
			return nil, err
		}
	}

	cd, err := cm.CreateContentData(ctx, obj.ContentType, obj.Data)
	if err != nil {
		log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
		return nil, err
	}

	switch v := cd.(type) {
	case *LessonData:
		//存在子内容，则返回子内容
		//if the content contains sub contents, return sub contents
		content, err := cm.ConvertContentObj(ctx, tx, obj, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrParseContentDataFailed
		}
		err = v.PrepareVersion(ctx)
		if err != nil {
			log.Error(ctx, "can't prepare version for sub contents", log.Err(err), log.Any("content", content))
			return nil, err
		}
		err = v.PrepareResult(ctx, tx, content, user, ignorePermissionFilter)
		if err != nil {
			log.Error(ctx, "can't get sub contents", log.Err(err), log.Any("content", content))
			return nil, err
		}
		ret := make([]*SubContentsWithName, 0)
		v.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
			if l.Material != nil {
				cd0, err := cm.CreateContentData(ctx, l.Material.ContentType, l.Material.Data)
				if err != nil {
					log.Error(ctx, "can't parse sub content data",
						log.Err(err),
						log.Any("lesson", l),
						log.Any("subContent", l.Material))
					return
				}
				ret = append(ret, &SubContentsWithName{
					ID:         l.Material.ID,
					Name:       l.Material.Name,
					Data:       cd0,
					OutcomeIDs: l.Material.Outcomes,
				})
			}
		})
		return ret, nil
	case *MaterialData:
		//若不存在子内容，则返回当前内容
		//if sub contents is not exists, return current content
		ret := []*SubContentsWithName{
			{
				ID:         cid,
				Name:       obj.Name,
				Data:       v,
				OutcomeIDs: cm.parseContentOutcomes(ctx, obj),
			},
		}
		return ret, nil
	case *AssetsData:
		//若不存在子内容，则返回当前内容
		//if sub contents is not exists, return current content
		ret := []*SubContentsWithName{
			{
				ID:         cid,
				Name:       obj.Name,
				Data:       v,
				OutcomeIDs: cm.parseContentOutcomes(ctx, obj),
			},
		}
		return ret, nil
	}

	return nil, ErrInvalidContentData
}

func (cm *ContentModel) GetContentNameByID(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentName, error) {
	cachedContent := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if cachedContent != nil {
		return &entity.ContentName{
			ID:          cid,
			Name:        cachedContent.Name,
			ContentType: cachedContent.ContentType,
			OutcomeIDs:  cachedContent.Outcomes,
			LatestID:    cachedContent.LatestID,
		}, nil
	}
	obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	latestID := obj.LatestID
	if latestID == "" {
		latestID = obj.ID
	}
	return &entity.ContentName{
		ID:          cid,
		Name:        obj.Name,
		ContentType: obj.ContentType,
		OutcomeIDs:  cm.parseContentOutcomes(ctx, obj),
		LatestID:    latestID,
	}, nil
}
func (cm *ContentModel) GetRawContentByIDListWithVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentWithVisibilitySettings, error) {
	contentList, err := cm.GetRawContentByIDList(ctx, tx, cids)
	if err != nil {
		return nil, err
	}
	visibilitySettings, err := cm.buildVisibilitySettingsMapByRawContent(ctx, contentList)
	if err != nil {
		log.Error(ctx, "buildVisibilitySettingsMapByRawContent failed", log.Err(err), log.Any("contentList", contentList))
		return nil, err
	}
	ret := make([]*entity.ContentWithVisibilitySettings, len(contentList))
	for i := range contentList {
		ret[i] = &entity.ContentWithVisibilitySettings{Content: *contentList[i]}
		ret[i].VisibilitySettings = visibilitySettings[ret[i].ID]
	}
	return ret, nil
}

func (cm *ContentModel) GetRawContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.Content, error) {
	obj, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.Strings("cids", cids))
		return nil, ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	return obj, nil
}

func (cm *ContentModel) GetContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	content := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if content == nil {
		log.Info(ctx, "Not cached content by id")
		obj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
		if err == dbo.ErrRecordNotFound {
			log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
			return nil, ErrNoContent
		}
		if err != nil {
			log.Error(ctx, "can't read contentdata", log.Err(err))
			return nil, err
		}
		content, err = cm.ConvertContentObj(ctx, tx, obj, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Err(err))
			return nil, ErrParseContentDataFailed
		}
		da.GetContentRedis().SaveContentCache(ctx, content)
	}
	log.Info(ctx, "pre fill content details", log.Any("content", content))
	return cm.fillContentDetails(ctx, tx, content, user)
}

func (cm *ContentModel) fillContentDetails(ctx context.Context, tx *dbo.DBContext, content *entity.ContentInfo, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	//补全相关内容
	//fill related data
	contentData, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		return nil, err
	}

	//TODO:For authed content => prepare result for latest version content => done
	err = contentData.PrepareVersion(ctx)
	if err != nil {
		log.Error(ctx, "can't update contentdata version for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	err = contentData.PrepareResult(ctx, tx, content, user, false)
	if err != nil {
		log.Error(ctx, "can't get contentdata for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	filledContentData, err := contentData.Marshal(ctx)
	if err != nil {
		log.Error(ctx, "can't marshal contentdata for details", log.Err(err))
		return nil, ErrParseContentDataDetailsFailed
	}
	content.Data = filledContentData

	contentWithDetails, err := cm.buildContentWithDetails(ctx, []*entity.ContentInfo{content}, true, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Err(err))
		return nil, ErrReadContentFailed
	}
	if len(contentWithDetails) < 1 {
		return &entity.ContentInfoWithDetails{
			ContentInfo: *content,
		}, nil
	}
	return contentWithDetails[0], nil
}

func (cm *ContentModel) GetContentNameByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentName, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]*entity.ContentName, 0)

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIDList(ctx, cids)
	for i := range cachedContent {
		resp = append(resp, &entity.ContentName{
			ID:          cachedContent[i].ID,
			Name:        cachedContent[i].Name,
			ContentType: cachedContent[i].ContentType,
			LatestID:    cachedContent[i].LatestID,
			OutcomeIDs:  cachedContent[i].Outcomes,
		})
	}
	if len(nid) < 1 {
		return resp, nil
	}

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, nid)
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		var latestID = data[i].LatestID
		if data[i].LatestID == "" {
			latestID = data[i].ID
		}
		resp = append(resp, &entity.ContentName{
			ID:          data[i].ID,
			Name:        data[i].Name,
			ContentType: data[i].ContentType,
			LatestID:    latestID,
			OutcomeIDs:  cm.parseContentOutcomes(ctx, data[i]),
		})
	}
	return resp, nil
}

func (cm *ContentModel) GetPastContentIDByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	data, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, ErrReadContentFailed
	}

	latestID := data.LatestID
	if data.LatestID == "" {
		latestID = data.ID
	}

	res, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		LatestID: latestID,
	})
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.String("latest id", data.LatestID))
		return nil, ErrReadContentFailed
	}
	resp := make([]string, len(res)+1)
	for i := range res {
		resp[i] = res[i].ID
	}
	resp[len(res)] = data.ID
	return utils.SliceDeduplication(resp), nil
}

func (cm *ContentModel) GetLatestContentIDByIDList(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]string, 0, len(cids))
	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for _, item := range data {
		if item.LatestID != "" {
			resp = append(resp, item.LatestID)
		} else {
			resp = append(resp, item.ID)
		}
	}
	return resp, nil
}

func (cm *ContentModel) IsContentsOperatorByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (bool, error) {
	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Strings("cids", cids), log.Err(err))
		return false, ErrReadContentFailed
	}
	for i := range data {
		if data[i].Author != user.UserID {
			return false, nil
		}
	}
	return true, nil
}

func (cm *ContentModel) GetContentByIDList(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	if len(cids) < 1 {
		return nil, nil
	}

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIDList(ctx, cids)
	//全在缓存中
	//all cached
	if len(nid) < 1 {
		contentWithDetails, err := cm.buildContentWithDetails(ctx, cachedContent, true, user)
		if err != nil {
			log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Err(err))
			return nil, ErrReadContentFailed
		}
		return contentWithDetails, nil
	}

	data, err := da.GetContentDA().GetContentByIDList(ctx, tx, nid)
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Strings("cids", cids), log.Err(err))
		return nil, ErrReadContentFailed
	}
	res, err := cm.BatchConvertContentObj(ctx, tx, data, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Any("data", data), log.Err(err))
		return nil, ErrReadContentFailed
	}
	res = append(res, cachedContent...)
	contentWithDetails, err := cm.buildContentWithDetails(ctx, res, true, user)
	if err != nil {
		log.Error(ctx, "can't parse contentdata", log.Strings("cids", cids), log.Err(err))
		return nil, ErrReadContentFailed
	}

	da.GetContentRedis().SaveContentCacheList(ctx, res)

	return contentWithDetails, nil
}

func (cm *ContentModel) SearchUserPrivateFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error) {
	//构造个人查询条件
	//construct query condition for private search
	condition.Author = user.UserID
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return 0, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	err = cm.filterRootPath(ctx, condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		return 0, nil, err
	}
	condition.VisibilitySettings = scope
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	condition.JoinUserIDList = searchUserIDs
	//cm.addUserCondition(ctx, &condition, user)

	//生成folder condition
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)

	log.Info(ctx, "search folder content", log.Any("condition", condition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	count, objs, err := da.GetContentDA().SearchFolderContent(ctx, tx, *cm.conditionRequestToCondition(*condition), folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("condition", condition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	ret := cm.convertFolderContent(ctx, objs, user)
	cm.fillFolderContent(ctx, ret, user)
	return count, ret, nil
}

func (cm *ContentModel) buildFolderConditionWithShowAll(ctx context.Context, op *entity.Operator, fCondition *da.FolderCondition) error {
	has, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, external.ShowAllFolder295)
	if err != nil {
		log.Error(ctx, "buildFolderConditionWithShowAll check permission failed",
			log.Err(err),
			log.Any("condition", fCondition),
			log.Any("op", op),
			log.String("permission", external.ShowAllFolder295))
		return err
	}
	fCondition.ShowEmptyFolder = entity.NullBool{Bool: has, Valid: true}
	return nil
}

func (cm *ContentModel) CountUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, error) {
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	err := cm.filterRootPath(ctx, condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		log.Warn(ctx, "filterRootPath failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, err
	}
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		log.Warn(ctx, "buildUserContentCondition failed", log.Err(err), log.Any("condition", condition), log.Any("searchUserIDs", searchUserIDs), log.String("uid", user.UserID))
		return 0, err
	}
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)

	log.Info(ctx, "count folder content", log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	total, err := da.GetContentDA().CountFolderContentUnsafe(ctx, tx, combineCondition, folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, ErrReadContentFailed
	}
	return total, nil
}

func (cm *ContentModel) SearchSimplifyContentInternal(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentInternalConditionRequest) (*entity.ContentSimplifiedList, error) {
	// search content by schedule id
	var schedule *entity.Schedule
	var scheduleReviews []*entity.ScheduleReview
	if condition.ScheduleID != "" {
		scheduleCondition := &entity.ScheduleQueryCondition{
			IDs: entity.NullStrings{
				Strings: []string{condition.ScheduleID},
				Valid:   true,
			},
		}
		schedules, err := GetScheduleModel().QueryUnsafe(ctx, scheduleCondition)
		if err != nil {
			log.Error(ctx, "GetScheduleModel().QueryUnsafe error",
				log.Err(err),
				log.Any("scheduleCondition", scheduleCondition))
			return nil, err
		}
		if len(schedules) == 0 {
			log.Error(ctx, "schedule not found",
				log.Err(err),
				log.Any("scheduleCondition", scheduleCondition))
			return nil, constant.ErrRecordNotFound
		}

		schedule = schedules[0]

		// review schedule query all student's content
		if schedule.IsReview {
			scheduleReviews, err = da.GetScheduleReviewDA().GetScheduleReviewsByScheduleID(ctx, dbo.MustGetDB(ctx), schedule.ID)
			if err != nil {
				log.Error(ctx, "search schedule review student content failed",
					log.Err(err),
					log.String("scheduleID", schedule.ID))
				return nil, err
			}
			for _, scheduleReview := range scheduleReviews {
				if scheduleReview.LiveLessonPlan != nil {
					for _, lessonMaterial := range scheduleReview.LiveLessonPlan.LessonMaterials {
						condition.IDs = append(condition.IDs, lessonMaterial.LessonMaterialID)
					}
				}
			}
		} else {
			// locked schedule query snapshot content
			if schedule.IsLockedLessonPlan() {
				for _, v := range schedule.LiveLessonPlan.LessonMaterials {
					condition.IDs = append(condition.IDs, v.LessonMaterialID)
				}
			} else {
				condition.PlanID = schedule.LessonPlanID
			}
		}

		// if schedule id exists, but there is no associated content, return empty
		if condition.IDs == nil && condition.PlanID == "" {
			return &entity.ContentSimplifiedList{
				Total:       0,
				ContentList: []*entity.ContentSimplified{},
			}, nil
		}
	}

	// get material ids from plan if condition contains plan id
	if condition.PlanID != "" {
		plan, err := da.GetContentDA().GetContentByID(ctx, tx, condition.PlanID)
		if err != nil {
			log.Error(ctx, "get plan failed", log.Err(err),
				log.String("plan_id", condition.PlanID),
				log.Any("condition", condition))
			return nil, err
		}
		if plan.LatestID != "" && plan.LatestID != plan.ID {
			plan, err = da.GetContentDA().GetContentByID(ctx, tx, plan.LatestID)
			if err != nil {
				log.Error(ctx, "get latest plan failed", log.Err(err),
					log.Any("plan", plan),
					log.Any("condition", condition))
				return nil, err
			}
		}
		if plan.ContentType != entity.ContentTypePlan {
			log.Error(ctx, "content data parse failed",
				log.Any("plan", plan))
			return nil, ErrInvalidContentType
		}
		cd, err := cm.CreateContentData(ctx, entity.ContentTypePlan, plan.Data)
		if err != nil {
			log.Error(ctx, "content data parse failed",
				log.Err(err),
				log.Any("plan", plan))
			return nil, err
		}
		err = cd.PrepareVersion(ctx)
		if err != nil {
			log.Error(ctx, "prepare material failed",
				log.Err(err),
				log.Any("cd", cd),
				log.Any("plan", plan),
				log.Any("condition", condition))
			return nil, err
		}
		planData, ok := cd.(*LessonData)
		if !ok {
			log.Error(ctx, "content data parse failed",
				log.Any("obj", cd),
				log.String("data", plan.Data),
			)
			return nil, ErrInvalidContentType
		}
		materialIDs := planData.SubContentIDs(ctx)
		//Add material IDs
		condition.IDs = append(condition.IDs, materialIDs...)
	}

	contentTypes := []int{condition.ContentType}
	if condition.ContentType == 0 {
		contentTypes = []int{entity.ContentTypeMaterial, entity.ContentTypePlan}
	}

	// Avoid pulling the full table
	if len(condition.IDs) == 0 &&
		condition.OrgID == "" &&
		len(contentTypes) == 0 &&
		condition.CreateAtLe == 0 &&
		condition.CreateAtGe == 0 {
		log.Error(ctx, "invalid search condition", log.Any("condition", condition))
		return nil, constant.ErrInvalidArgs
	}

	cdt := &da.ContentCondition{
		IDS: entity.NullStrings{
			Valid:   condition.IDs != nil,
			Strings: condition.IDs,
		},
		Org:            condition.OrgID,
		ContentType:    contentTypes,
		DataSourceID:   condition.DataSourceID,
		CreateAtLe:     condition.CreateAtLe,
		CreateAtGe:     condition.CreateAtGe,
		IncludeDeleted: true,
	}

	total, data, err := da.GetContentDA().SearchContent(ctx, tx, cdt)
	if err != nil {
		log.Error(ctx, "search content internal failed",
			log.Err(err),
			log.Any("condition", cdt))
		return nil, err
	}
	contentIDs := make([]string, len(data))
	res := make([]*entity.ContentSimplified, len(data))
	for i := range data {
		res[i] = data[i].ToContentSimplified()
		contentIDs[i] = data[i].ID
	}

	var studentContentMap []*entity.ScheduleStudentContent
	// get schedule review student content map
	if schedule != nil && schedule.IsReview {
		for _, scheduleReview := range scheduleReviews {
			studentContentIDs := []string{}
			if scheduleReview.LiveLessonPlan != nil {
				for _, lessonMaterial := range scheduleReview.LiveLessonPlan.LessonMaterials {
					if utils.ContainsString(contentIDs, lessonMaterial.LessonMaterialID) {
						studentContentIDs = append(studentContentIDs, lessonMaterial.LessonMaterialID)
					}
				}
			}
			studentContentMap = append(studentContentMap, &entity.ScheduleStudentContent{
				StudentID:  scheduleReview.StudentID,
				ContentIDs: studentContentIDs,
			})
		}
	}

	return &entity.ContentSimplifiedList{
		Total:             total,
		ContentList:       res,
		StudentContentMap: studentContentMap,
	}, nil
}

func (cm *ContentModel) SearchUserFolderContentSlim(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error) {
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	err := cm.filterRootPath(ctx, condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		log.Warn(ctx, "filterRootPath failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		log.Warn(ctx, "buildUserContentCondition failed", log.Err(err), log.Any("condition", condition), log.Any("searchUserIDs", searchUserIDs), log.String("uid", user.UserID))
		return 0, nil, err
	}

	foldPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, user, []external.PermissionName{
		external.CreateFolder289,
		external.ShowAllFolder295,
	})
	if err != nil {
		log.Error(ctx, "hasFolderPermission failed",
			log.Err(err),
			log.Any("condition", condition),
			log.Any("user", user))
		return 0, nil, err
	}
	folderCondition := cm.buildFolderConditionWithPermission(ctx, user, condition, searchUserIDs, foldPermission)

	log.Info(ctx, "search folder content", log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	count, objs, err := da.GetContentDA().SearchFolderContentUnsafe(ctx, tx, combineCondition, folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	ret := cm.convertFolderContent(ctx, objs, user)
	cm.fillFolderContentPermissionSlim(ctx, user, ret, foldPermission)
	return count, ret, nil
}

func (cm *ContentModel) SearchUserFolderContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.FolderContentData, error) {
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	err := cm.filterRootPath(ctx, condition, entity.OwnerTypeOrganization, user)
	if err != nil {
		log.Warn(ctx, "filterRootPath failed", log.Err(err), log.Any("condition", condition), log.String("uid", user.UserID))
		return 0, nil, err
	}
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		log.Warn(ctx, "buildUserContentCondition failed", log.Err(err), log.Any("condition", condition), log.Any("searchUserIDs", searchUserIDs), log.String("uid", user.UserID))
		return 0, nil, err
	}
	folderCondition := cm.buildFolderCondition(ctx, condition, searchUserIDs, user)
	if len(condition.PublishStatus) == 1 && condition.PublishStatus[0] == entity.ContentStatusPublished {
		err = cm.buildFolderConditionWithShowAll(ctx, user, folderCondition)
		if err != nil {
			log.Warn(ctx, "SearchUserFolderContent append show all failed",
				log.Any("op", user),
				log.Any("condition", condition))
			return 0, nil, err
		}
	}

	log.Info(ctx, "search folder content", log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
	count, objs, err := da.GetContentDA().SearchFolderContentUnsafe(ctx, tx, combineCondition, folderCondition)
	if err != nil {
		log.Error(ctx, "can't read folder content", log.Err(err), log.Any("combineCondition", combineCondition), log.Any("folderCondition", folderCondition), log.String("uid", user.UserID))
		return 0, nil, ErrReadContentFailed
	}
	ret := cm.convertFolderContent(ctx, objs, user)
	cm.fillFolderContent(ctx, ret, user)
	return count, ret, nil
}
func (cm *ContentModel) SearchUserContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	//where, params := combineCondition.GetConditions()
	//logger.WithContext(ctx).WithField("subject", "content").Infof("Combine condition: %#v, params: %#v", where, params)
	searchUserIDs := cm.getRelatedUserID(ctx, condition.Name, user)
	combineCondition, err := cm.buildUserContentCondition(ctx, tx, condition, searchUserIDs, user)
	if err != nil {
		return 0, nil, err
	}
	return cm.searchContentUnsafe(ctx, tx, combineCondition, user)
}

func (cm *ContentModel) SearchUserPrivateContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.Author = user.UserID
	condition.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition.PublishStatus)
	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return 0, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition.VisibilitySettings = scope

	cm.addUserCondition(ctx, condition, user)
	return cm.searchContent(ctx, tx, condition, user)
}

func (cm *ContentModel) ListPendingContent(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, user *entity.Operator) (int, []*entity.ContentInfoWithDetails, error) {
	condition.PublishStatus = []string{entity.ContentStatusPending}
	//scope, err := cm.ListVisibleScopes(ctx, visiblePermissionPending, user)
	//if err != nil {
	//	return 0, nil, err
	//}
	//if len(scope) == 0 {
	//	log.Info(ctx, "no valid private scope", log.Strings("scopes", scope), log.Any("user", user))
	//	scope = []string{constant.NoSearchItem}
	//}
	//condition.VisibilitySettings = scope

	cm.addUserCondition(ctx, condition, user)
	return cm.searchContent(ctx, tx, condition, user)
}

func (cm *ContentModel) addUserCondition(ctx context.Context, condition *entity.ContentConditionRequest, user *entity.Operator) {
	condition.JoinUserIDList = cm.getRelatedUserID(ctx, condition.Name, user)
}

func (cm *ContentModel) getRelatedUserID(ctx context.Context, keyword string, user *entity.Operator) []string {
	if keyword == "" {
		return nil
	}
	users, err := external.GetUserServiceProvider().Query(ctx, user, user.OrgID, keyword)
	if err != nil {
		log.Warn(ctx, "get user info failed", log.Err(err), log.String("keyword", keyword), log.Any("user", user))
		return nil
	}
	if len(users) < 1 {
		log.Info(ctx, "user info not found in keywords", log.Err(err), log.String("keyword", keyword), log.String("userID", user.UserID), log.String("orgID", user.OrgID))
		return nil
	}
	ids := make([]string, len(users))
	for i := range users {
		ids[i] = users[i].ID
	}
	return ids
}

func (cm *ContentModel) refreshContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	alreadyScopes, err := da.GetContentDA().GetContentVisibilitySettings(ctx, tx, cid)
	if err != nil {
		log.Error(ctx,
			"GetContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("scope", scope))
		return err
	}
	pendingAddScopes := cm.checkDiff(ctx, alreadyScopes, scope)
	pendingDeleteScopes := cm.checkDiff(ctx, scope, alreadyScopes)

	log.Info(ctx,
		"BatchRefreshContentVisibilitySettings",
		log.String("cid", cid),
		log.Strings("alreadyScopes", alreadyScopes),
		log.Strings("scope", scope),
		log.Strings("pendingDeleteScopes", pendingDeleteScopes),
		log.Strings("pendingAddScopes", pendingAddScopes))

	err = da.GetContentDA().BatchDeleteContentVisibilitySettings(ctx, tx, cid, pendingDeleteScopes)
	if err != nil {
		log.Error(ctx,
			"BatchDeleteContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("alreadyScopes", alreadyScopes),
			log.Strings("scope", scope),
			log.Strings("pendingDeleteScopes", pendingDeleteScopes))
		return err
	}

	err = da.GetContentDA().BatchCreateContentVisibilitySettings(ctx, tx, cid, pendingAddScopes)
	if err != nil {
		log.Error(ctx,
			"BatchCreateContentVisibilitySettings failed",
			log.Err(err),
			log.String("cid", cid),
			log.Strings("alreadyScopes", alreadyScopes),
			log.Strings("scope", scope),
			log.Strings("pendingAddScopes", pendingAddScopes))
		return err
	}
	return nil
}

func (cm *ContentModel) checkDiff(ctx context.Context, base []string, compare []string) []string {
	sourceMap := make(map[string]bool)
	ret := make([]string, 0)
	for i := range base {
		sourceMap[base[i]] = true
	}
	for i := range compare {
		_, exist := sourceMap[compare[i]]
		if !exist {
			ret = append(ret, compare[i])
		}
	}
	return ret
}

func (cm *ContentModel) insertContentVisibilitySettings(ctx context.Context, tx *dbo.DBContext, cid string, scope []string) error {
	//Insert visibility settings
	err := da.GetContentDA().BatchCreateContentVisibilitySettings(ctx, tx, cid, scope)
	if err != nil {
		log.Error(ctx,
			"BatchCreateContentVisibilitySettings failed",
			log.Err(err),
			log.Strings("scope", scope),
			log.String("cid", cid))
		return err
	}
	return nil
}

func (cm *ContentModel) getVisibleContentOutcomeByIDs(ctx context.Context, tx *dbo.DBContext, cids []string) ([]string, error) {
	//get latest content ids
	newCids, err := cm.GetLatestContentIDByIDList(ctx, tx, cids)
	if err != nil {
		log.Error(ctx, "can't get content latest id", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	//get content objects
	contents, err := da.GetContentDA().GetContentByIDList(ctx, tx, newCids)
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.Strings("newCids", newCids))
		return nil, err
	}

	//collect outcomes
	ret := make([]string, 0)
	for i := range contents {
		if contents[i].Outcomes == "" {
			continue
		}
		outcomes := strings.Split(contents[i].Outcomes, constant.StringArraySeparator)
		for i := range outcomes {
			if outcomes[i] != "" {
				ret = append(ret, outcomes[i])
			}
		}
	}
	ret = utils.SliceDeduplication(ret)
	return ret, nil
}

func (cm *ContentModel) GetVisibleContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	if content.LatestID != "" {
		content, err = da.GetContentDA().GetContentByID(ctx, tx, content.LatestID)
		if err != nil {
			log.Error(ctx, "can't get latest content", log.Err(err), log.String("cid", cid))
			return nil, err
		}
	}
	ret := make([]string, 0)
	//if content is a plan, collect outcomes from materials
	if content.ContentType == entity.ContentTypePlan {
		contentData, err := cm.CreateContentData(ctx, entity.ContentTypePlan, content.Data)
		if err != nil {
			log.Error(ctx, "can't parse content data",
				log.Err(err),
				log.String("cid", cid),
				log.String("data", content.Data))
			return nil, err
		}
		contentIDList := contentData.(*LessonData).SubContentIDs(ctx)
		outcomes, err := cm.getVisibleContentOutcomeByIDs(ctx, tx, contentIDList)
		if err != nil {
			log.Error(ctx, "can't get outcomes from materials",
				log.Err(err),
				log.Strings("contentIDList", contentIDList))
			return nil, err
		}
		ret = append(ret, outcomes...)
	}

	if content.Outcomes == "" {
		return ret, nil
	}
	outcomes := strings.Split(content.Outcomes, constant.StringArraySeparator)
	for i := range outcomes {
		if outcomes[i] != "" {
			ret = append(ret, outcomes[i])
		}
	}
	ret = utils.SliceDeduplication(ret)

	return ret, nil
}

func (cm *ContentModel) GetContentOutcomeByID(ctx context.Context, tx *dbo.DBContext, cid string) ([]string, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	ret := cm.parseContentOutcomes(ctx, content)
	//if content is a lesson, add materials outcomes
	if content.ContentType == entity.ContentTypePlan {
		data, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
		if err != nil {
			log.Error(ctx, "parse content data failed",
				log.Err(err),
				log.Any("content", content))
			return nil, err
		}
		materialIDs := data.SubContentIDs(ctx)
		if len(materialIDs) > 0 {
			materials, err := da.GetContentDA().GetContentByIDList(ctx, tx, materialIDs)
			if err != nil {
				log.Error(ctx, "parse content data failed",
					log.Err(err),
					log.Any("content", content))
				return nil, err
			}

			for i := range materials {
				outcomes := cm.parseContentOutcomes(ctx, materials[i])
				if len(outcomes) > 0 {
					ret = append(ret, outcomes...)
				}
			}
		}
	}

	return ret, nil
}

func (cm *ContentModel) ContentDataCount(ctx context.Context, tx *dbo.DBContext, cid string) (*entity.ContentStatisticsInfo, error) {
	content, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err != nil {
		log.Error(ctx, "can't get content", log.Err(err), log.String("cid", cid))
		return nil, err
	}
	if content.ContentType != entity.ContentTypePlan {
		log.Error(ctx, "invalid content type", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)))
		return nil, ErrInvalidContentType
	}

	cd, err := cm.CreateContentData(ctx, content.ContentType, content.Data)
	if err != nil {
		log.Error(ctx, "can't parse content data", log.Err(err), log.String("cid", cid), log.Int("contentType", int(content.ContentType)), log.String("data", content.Data))
		return nil, err
	}
	subContentIDs := cd.SubContentIDs(ctx)

	subContents, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IDS: entity.NullStrings{
			Strings: subContentIDs,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "search data failed", log.Err(err), log.String("cid", cid),
			log.Int("contentType", int(content.ContentType)),
			log.String("data", content.Data),
			log.Strings("subContentIDs", subContentIDs))
		return nil, err
	}

	identityOutComes := make(map[string]bool)
	outcomesCount := 0
	for i := range subContents {
		if subContents[i].Outcomes == "" {
			continue
		}
		subOutcomes := strings.Split(subContents[i].Outcomes, constant.StringArraySeparator)
		for j := range subOutcomes {
			_, ok := identityOutComes[subOutcomes[j]]
			if !ok {
				outcomesCount++
			}
			identityOutComes[subOutcomes[j]] = true
		}
	}
	return &entity.ContentStatisticsInfo{
		SubContentCount: len(subContentIDs),
		OutcomesCount:   outcomesCount,
	}, nil
}

func (cm *ContentModel) GetVisibleContentByID(ctx context.Context, tx *dbo.DBContext, cid string, user *entity.Operator) (*entity.ContentInfoWithDetails, error) {
	var err error
	log.Info(ctx, "GetVisibleContentByID")
	cachedContent := da.GetContentRedis().GetContentCacheByID(ctx, cid)
	if cachedContent != nil {
		if cachedContent.LatestID == "" {
			log.Info(ctx, "Cached latest content")
			return cm.fillContentDetails(ctx, tx, cachedContent, user)
		} else {
			log.Info(ctx, "Cached content not latest")
			return cm.GetContentByID(ctx, tx, cachedContent.LatestID, user)
		}
	}
	log.Info(ctx, "Not cached latest content")

	contentObj, err := da.GetContentDA().GetContentByID(ctx, tx, cid)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("cid", cid), log.String("uid", user.UserID))
		return nil, ErrNoContent
	}
	if err != nil {
		log.Error(ctx, "can't read contentdata", log.Err(err))
		return nil, err
	}
	if contentObj.LatestID == "" {
		return cm.GetContentByID(ctx, tx, cid, user)
	}
	return cm.GetContentByID(ctx, tx, contentObj.LatestID, user)
}

func (cm *ContentModel) filterInvisiblePublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	if len(status) < 1 {
		return []string{
			entity.ContentStatusPublished,
			entity.ContentStatusPending,
			entity.ContentStatusDraft,
			entity.ContentStatusRejected,
		}
	}

	for i := range status {
		if status[i] != entity.ContentStatusAttachment &&
			status[i] != entity.ContentStatusHidden {
			newStatus = append(newStatus, status[i])
		}
	}
	return newStatus
}

func (cm *ContentModel) parseContentOutcomes(ctx context.Context, content *entity.Content) []string {
	if content.Outcomes == "" {
		return nil
	}
	outcomes := strings.Split(content.Outcomes, constant.StringArraySeparator)
	ret := make([]string, 0)
	for i := range outcomes {
		if outcomes[i] != "" {
			ret = append(ret, outcomes[i])
		}
	}
	return ret
}

func (cm *ContentModel) filterRootPath(ctx context.Context, condition *entity.ContentConditionRequest, ownerType entity.OwnerType, operator *entity.Operator) error {
	if condition.DirPath != "" {
		return nil
	}

	//isAssets := false
	//for i := range condition.ContentType {
	//	if entity.NewContentType(condition.ContentType[i]).IsAsset() {
	//		isAssets = true
	//		break
	//	}
	//}
	//
	//if isAssets {
	//	root, err := GetFolderModel().GetRootFolder(ctx, entity.RootAssetsFolderName, ownerType, operator)
	//	if err != nil{
	//		log.Error(ctx, "can't get root folder", log.Err(err), log.Any("ownerType", ownerType), log.Any("operator", operator), log.String("partition", string(entity.RootAssetsFolderName)))
	//		return err
	//	}
	//	condition.DirPath = string(root.ChildrenPath())
	//	return nil
	//}
	//
	//root, err := GetFolderModel().GetRootFolder(ctx, entity.RootMaterialsAndPlansFolderName, ownerType, operator)
	//if err != nil{
	//	log.Error(ctx, "can't get root folder", log.Err(err), log.Any("ownerType", ownerType), log.Any("operator", operator), log.String("partition", string(entity.RootMaterialsAndPlansFolderName)))
	//	return err
	//}
	//condition.DirPath = string(root.ChildrenPath())
	condition.DirPath = constant.FolderRootPath
	return nil
}

func (cm *ContentModel) filterPublishedPublishStatus(ctx context.Context, status []string) []string {
	newStatus := make([]string, 0)
	for i := range status {
		if status[i] == entity.ContentStatusPublished ||
			status[i] == entity.ContentStatusArchive {
			newStatus = append(newStatus, status[i])
		}
	}
	if len(newStatus) < 1 {
		return []string{
			entity.ContentStatusPublished,
		}
	}
	return newStatus
}

func (cm *ContentModel) buildUserContentCondition(ctx context.Context, tx *dbo.DBContext, condition *entity.ContentConditionRequest, searchUserIDs []string, user *entity.Operator) (dbo.Conditions, error) {

	err := cm.convertAuthedOrgIDToParentsPath(ctx, tx, condition)
	if err != nil {
		return nil, err
	}

	condition1 := *condition
	condition2 := *condition

	//condition1 private
	condition1.Author = user.UserID
	condition1.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition1.PublishStatus)

	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid private scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition1.VisibilitySettings = scope

	//condition2 others
	condition2.PublishStatus = cm.filterPublishedPublishStatus(ctx, condition2.PublishStatus)

	//filter visible
	if len(condition.ContentType) == 1 && condition.ContentType[0] == entity.ContentTypeAssets {
		condition2.VisibilitySettings = []string{user.OrgID}
	}

	//condition2.Scope = scopes
	condition1.JoinUserIDList = searchUserIDs
	condition2.JoinUserIDList = searchUserIDs

	//condition1.UseJoinForVisibilitySettings = true
	//condition2.UseJoinForVisibilitySettings = true

	if condition.PublishedQueryMode == entity.PublishedQueryModeOnlyOwner {
		//The user only has the permission to query his own
		return cm.conditionRequestToCondition(condition1), nil
	} else if condition.PublishedQueryMode == entity.PublishedQueryModeOnlyOthers {
		//The user only has the permission to query others
		return cm.conditionRequestToCondition(condition2), nil
	} else if condition.PublishedQueryMode == entity.PublishedQueryModeNone {
		log.Error(ctx, "no valid private scope",
			log.Err(err),
			log.Strings("scopes", scope),
			log.Any("condition", condition),
			log.Any("user", user))
		return nil, ErrNoPermissionToQuery
	}

	combineCondition := &da.CombineConditions{
		SourceCondition: cm.conditionRequestToCondition(condition1),
		TargetCondition: cm.conditionRequestToCondition(condition2),
	}
	return combineCondition, nil
}

func (cm *ContentModel) buildTreeContentCondition(ctx context.Context, tx *dbo.DBContext,
	condition *entity.ContentConditionRequest, searchUserIDs []string, user *entity.Operator) (
	myContentCondition *da.ContentCondition, otherContentCondition *da.ContentCondition, err error) {

	err = cm.convertAuthedOrgIDToParentsPath(ctx, tx, condition)
	if err != nil {
		return nil, nil, err
	}

	condition1 := *condition
	condition2 := *condition

	//condition1 private
	condition1.Author = user.UserID
	condition1.PublishStatus = cm.filterInvisiblePublishStatus(ctx, condition1.PublishStatus)

	scope, err := cm.listAllScopes(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	if len(scope) == 0 {
		log.Info(ctx, "no valid private scope", log.Strings("scopes", scope), log.Any("user", user))
		scope = []string{constant.NoSearchItem}
	}
	condition1.VisibilitySettings = scope

	//condition2 others
	condition2.PublishStatus = cm.filterPublishedPublishStatus(ctx, condition2.PublishStatus)

	//filter visible
	if len(condition.ContentType) == 1 && condition.ContentType[0] == entity.ContentTypeAssets {
		condition2.VisibilitySettings = []string{user.OrgID}
	}

	//condition2.Scope = scopes
	condition1.JoinUserIDList = searchUserIDs
	condition2.JoinUserIDList = searchUserIDs

	if condition.PublishedQueryMode == entity.PublishedQueryModeOnlyOwner {
		//The user only has the permission to query his own
		return cm.conditionRequestToCondition(condition1), nil, nil
	} else if condition.PublishedQueryMode == entity.PublishedQueryModeOnlyOthers {
		//The user only has the permission to query others
		return cm.conditionRequestToCondition(condition2), nil, nil
	} else if condition.PublishedQueryMode == entity.PublishedQueryModeNone {
		log.Error(ctx, "no valid private scope",
			log.Err(err),
			log.Strings("scopes", scope),
			log.Any("condition", condition),
			log.Any("user", user))
		return nil, nil, ErrNoPermissionToQuery
	}
	return cm.conditionRequestToCondition(condition1), cm.conditionRequestToCondition(condition2), nil
}

func (cm *ContentModel) checkPublishContentChildren(ctx context.Context, c *entity.Content, children []*entity.Content) error {
	//TODO: To implement, check publish scope
	for i := range children {
		if children[i].PublishStatus != entity.ContentStatusPublished &&
			children[i].PublishStatus != entity.ContentStatusHidden {
			log.Warn(ctx, "check children status failed", log.Any("content", children[i]))
			return ErrPlanHasArchivedMaterials
		}
	}
	//TODO:For authed content => update check for authed content list

	return nil
}

func (cm *ContentModel) buildVisibilitySettingsMap(ctx context.Context, contentList []*entity.ContentInfo) (map[string][]string, error) {
	contentIDs := make([]string, len(contentList))
	for i := range contentList {
		contentIDs[i] = contentList[i].ID
	}

	visibilitySettings, err := da.GetContentDA().SearchContentVisibilitySettings(ctx, dbo.MustGetDB(ctx), &da.ContentVisibilitySettingsCondition{
		ContentIDs: contentIDs,
	})
	if err != nil {
		log.Error(ctx, "GetContentVisibilitySettings failed",
			log.Err(err),
			log.Any("contentList", contentList))
		return nil, err
	}

	ret := make(map[string][]string)
	for i := range visibilitySettings {
		ret[visibilitySettings[i].ContentID] = append(ret[visibilitySettings[i].ContentID], visibilitySettings[i].VisibilitySetting)
	}
	return ret, nil
}

func (cm *ContentModel) buildVisibilitySettingsMapByRawContent(ctx context.Context, contentList []*entity.Content) (map[string][]string, error) {
	contentIDs := make([]string, len(contentList))
	for i := range contentList {
		contentIDs[i] = contentList[i].ID
	}

	visibilitySettings, err := da.GetContentDA().SearchContentVisibilitySettings(ctx, dbo.MustGetDB(ctx), &da.ContentVisibilitySettingsCondition{
		ContentIDs: contentIDs,
	})
	if err != nil {
		log.Error(ctx, "GetContentVisibilitySettings failed",
			log.Err(err),
			log.Any("contentList", contentList))
		return nil, err
	}

	ret := make(map[string][]string)
	for i := range visibilitySettings {
		ret[visibilitySettings[i].ContentID] = append(ret[visibilitySettings[i].ContentID], visibilitySettings[i].VisibilitySetting)
	}
	return ret, nil
}

func (cm *ContentModel) doCreateContentProperties(ctx context.Context, tx *dbo.DBContext, c entity.ContentProperties, refresh bool) error {
	propertyDA := da.GetContentPropertyDA()

	if refresh {
		//delete old properties
		err := propertyDA.CleanByContentID(ctx, tx, c.ContentID)
		if err != nil {
			log.Error(ctx, "CleanByContentID failed",
				log.Err(err),
				log.String("ContentID", c.ContentID),
				log.Any("data", c))
			return err
		}
	}
	size := len(c.Grade) + len(c.Subject) + len(c.Age) + len(c.Category) + len(c.SubCategory)
	properties := make([]*entity.ContentProperty, size+1)
	index := 0
	for i := range c.Grade {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeGrade,
			ContentID:    c.ContentID,
			PropertyID:   c.Grade[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Subject {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubject,
			ContentID:    c.ContentID,
			PropertyID:   c.Subject[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Age {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeAge,
			ContentID:    c.ContentID,
			PropertyID:   c.Age[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.Category {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeCategory,
			ContentID:    c.ContentID,
			PropertyID:   c.Category[i],
			Sequence:     i,
		}
		index++
	}
	for i := range c.SubCategory {
		properties[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubCategory,
			ContentID:    c.ContentID,
			PropertyID:   c.SubCategory[i],
			Sequence:     i,
		}
		index++
	}
	properties[index] = &entity.ContentProperty{
		PropertyType: entity.ContentPropertyTypeProgram,
		ContentID:    c.ContentID,
		PropertyID:   c.Program,
		Sequence:     0,
	}

	err := propertyDA.BatchAdd(ctx, tx, properties)
	if err != nil {
		log.Error(ctx, "BatchAdd failed",
			log.Err(err),
			log.String("ContentID", c.ContentID),
			log.Any("properties", properties),
			log.Any("data", c))
		return err
	}
	return nil
}

func (cm *ContentModel) buildContentWithDetails(ctx context.Context, contentList []*entity.ContentInfo, includeOutcomes bool, user *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	orgName := ""
	orgProvider := external.GetOrganizationServiceProvider()
	orgs, err := orgProvider.BatchGet(ctx, user, []string{user.OrgID})
	if err != nil || len(orgs) < 1 {
		log.Error(ctx, "can't get org info", log.Err(err))
	} else {
		if orgs[0].Valid {
			orgName = orgs[0].Name
		} else {
			log.Warn(ctx, "invalid value", log.String("org_id", user.OrgID))
		}
	}

	programNameMap := make(map[string]string)
	subjectNameMap := make(map[string]string)
	developmentalNameMap := make(map[string]string)
	skillsNameMap := make(map[string]string)
	ageNameMap := make(map[string]string)
	gradeNameMap := make(map[string]string)
	publishScopeNameMap := make(map[string]string)
	lessonTypeNameMap := make(map[string]string)
	userNameMap := make(map[string]string)

	programIDs := make([]string, 0)
	subjectIDs := make([]string, 0)
	developmentalIDs := make([]string, 0)
	skillsIDs := make([]string, 0)
	ageIDs := make([]string, 0)
	gradeIDs := make([]string, 0)
	scopeIDs := make([]string, 0)
	lessonTypeIDs := make([]string, 0)
	userIDs := make([]string, 0)

	visibilitySettingsMap, err := cm.buildVisibilitySettingsMap(ctx, contentList)
	if err != nil {
		log.Error(ctx, "buildVisibilitySettingsMap failed",
			log.Err(err),
			log.Any("contentList", contentList))
		return nil, err
	}

	for i := range contentList {
		if contentList[i].Program != "" {
			programIDs = append(programIDs, contentList[i].Program)
		}
		subjectIDs = append(subjectIDs, contentList[i].Subject...)
		developmentalIDs = append(developmentalIDs, contentList[i].Category...)
		skillsIDs = append(skillsIDs, contentList[i].SubCategory...)
		ageIDs = append(ageIDs, contentList[i].Age...)
		gradeIDs = append(gradeIDs, contentList[i].Grade...)

		scopeIDs = append(scopeIDs, visibilitySettingsMap[contentList[i].ID]...)
		lessonTypeIDs = append(lessonTypeIDs, contentList[i].LessonType)
		userIDs = append(userIDs, contentList[i].Author)
		userIDs = append(userIDs, contentList[i].Creator)
	}

	//Users
	userNameMap, err = external.GetUserServiceProvider().BatchGetNameMap(ctx, user, userIDs)
	if err != nil {
		log.Error(ctx, "can't get user info", log.Err(err), log.Strings("ids", userIDs))
	}

	//LessonType
	lessonTypes, err := GetLessonTypeModel().Query(ctx, &da.LessonTypeCondition{
		IDs: entity.NullStrings{
			Strings: lessonTypeIDs,
			Valid:   len(lessonTypeIDs) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "can't get lesson type info", log.Err(err))
	} else {
		for i := range lessonTypes {
			lessonTypeNameMap[lessonTypes[i].ID] = lessonTypes[i].Name
		}
	}

	//Program
	programNameMap, err = external.GetProgramServiceProvider().BatchGetNameMap(ctx, user, programIDs)
	if err != nil {
		log.Error(ctx, "can't get programs", log.Err(err), log.Strings("ids", programIDs))
	}

	//Subjects
	subjectNameMap, err = external.GetSubjectServiceProvider().BatchGetNameMap(ctx, user, subjectIDs)
	if err != nil {
		log.Error(ctx, "can't get subjects info", log.Err(err))
	}

	//developmental
	developmentalNameMap, err = external.GetCategoryServiceProvider().BatchGetNameMap(ctx, user, developmentalIDs)
	if err != nil {
		log.Error(ctx, "can't get category info", log.Err(err), log.Strings("ids", developmentalIDs))
	}

	//scope
	//TODO:change to get org name
	publishScopeNameMap, err = external.GetOrganizationServiceProvider().GetNameMapByOrganizationOrSchool(ctx, user, scopeIDs)
	if err != nil {
		log.Error(ctx, "can't get publish scope info", log.Strings("scope", scopeIDs), log.Err(err))
	}

	//skill
	skillsNameMap, err = external.GetSubCategoryServiceProvider().BatchGetNameMap(ctx, user, skillsIDs)
	if err != nil {
		log.Error(ctx, "can't get skills info", log.Strings("skillsIDs", skillsIDs), log.Err(err))
	}

	//age
	ageNameMap, err = external.GetAgeServiceProvider().BatchGetNameMap(ctx, user, ageIDs)
	if err != nil {
		log.Error(ctx, "can't get age info", log.Strings("ageIDs", ageIDs), log.Err(err))
	}

	//grade
	gradeNameMap, err = external.GetGradeServiceProvider().BatchGetNameMap(ctx, user, gradeIDs)
	if err != nil {
		log.Error(ctx, "can't get grade info", log.Strings("gradeIDs", gradeIDs), log.Err(err))
	}

	var outcomeDictionary map[string]*entity.Outcome
	var outcomeSortIDs []string
	if includeOutcomes {
		outcomeMap := make(map[string]bool)
		outcomeIDs := make([]string, 0)
		for _, content := range contentList {
			for _, id := range content.Outcomes {
				if outcomeMap[id] {
					continue
				}

				outcomeMap[id] = true
				outcomeIDs = append(outcomeIDs, id)
			}
		}

		outcomeDictionary, outcomeSortIDs, err = GetOutcomeModel().GetLatestOutcomes(ctx, user, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			log.Warn(ctx, "get latest outcomes entity failed",
				log.Err(err),
				log.Strings("ids", outcomeIDs),
				log.String("uid", user.UserID))
		}
	}

	contentDetailsList := make([]*entity.ContentInfoWithDetails, len(contentList))
	for i := range contentList {
		subjectNames := make([]string, len(contentList[i].Subject))
		developmentalNames := make([]string, len(contentList[i].Category))
		skillsNames := make([]string, len(contentList[i].SubCategory))
		ageNames := make([]string, len(contentList[i].Age))
		gradeNames := make([]string, len(contentList[i].Grade))

		for j := range contentList[i].Subject {
			subjectNames[j] = subjectNameMap[contentList[i].Subject[j]]
		}
		for j := range contentList[i].Category {
			developmentalNames[j] = developmentalNameMap[contentList[i].Category[j]]
		}
		for j := range contentList[i].SubCategory {
			skillsNames[j] = skillsNameMap[contentList[i].SubCategory[j]]
		}
		for j := range contentList[i].Age {
			ageNames[j] = ageNameMap[contentList[i].Age[j]]
		}
		for j := range contentList[i].Grade {
			gradeNames[j] = gradeNameMap[contentList[i].Grade[j]]
		}

		contentList[i].PublishScope = visibilitySettingsMap[contentList[i].ID]
		publishScopeNames := make([]string, len(contentList[i].PublishScope))

		for j := range contentList[i].PublishScope {
			publishScopeNames[j] = publishScopeNameMap[contentList[i].PublishScope[j]]
		}
		contentList[i].AuthorName = userNameMap[contentList[i].Author]
		contentDetailsList[i] = &entity.ContentInfoWithDetails{
			ContentInfo:      *contentList[i],
			ContentTypeName:  contentList[i].ContentType.Name(),
			ProgramName:      programNameMap[contentList[i].Program],
			SubjectName:      subjectNames,
			CategoryName:     developmentalNames,
			SubCategoryName:  skillsNames,
			AgeName:          ageNames,
			GradeName:        gradeNames,
			LessonTypeName:   lessonTypeNameMap[contentList[i].LessonType],
			PublishScopeName: publishScopeNames,
			OrgName:          orgName,
			//AuthorName:        userNameMap[contentList[i].Author],
			CreatorName:     userNameMap[contentList[i].Creator],
			OutcomeEntities: []*entity.Outcome{},
			IsMine:          contentList[i].Author == user.UserID,
		}

		if includeOutcomes {
			for _, id := range outcomeSortIDs {
				if utils.ContainsString(contentList[i].Outcomes, id) && outcomeDictionary != nil {
					contentDetailsList[i].OutcomeEntities = append(contentDetailsList[i].OutcomeEntities, outcomeDictionary[id])
				}
			}
		}
	}

	//fill content permission info
	cm.fillContentPermission(ctx, contentDetailsList, user)

	log.Info(ctx, "build content detail list successfully", log.Any("content", contentDetailsList))

	return contentDetailsList, nil
}

func (cm *ContentModel) buildContentWithDetailsForSearchContent(ctx context.Context, contentList []*entity.ContentInfo, op *entity.Operator) ([]*entity.ContentInfoWithDetails, error) {
	userNameMap := make(map[string]string)
	userIDs := make([]string, 0)

	for i := range contentList {
		userIDs = append(userIDs, contentList[i].Author)
		userIDs = append(userIDs, contentList[i].Creator)
	}

	// map[userID]userName
	userNameMap, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "can't get user info", log.Err(err), log.Strings("ids", userIDs))
	}
	contentDetailsList := make([]*entity.ContentInfoWithDetails, len(contentList))
	for i := range contentList {
		contentList[i].AuthorName = userNameMap[contentList[i].Author]
		contentDetailsList[i] = &entity.ContentInfoWithDetails{
			ContentInfo:     *contentList[i],
			ContentTypeName: contentList[i].ContentType.Name(),
			IsMine:          contentList[i].Author == op.UserID,
		}
	}

	//fill content permission info
	cm.fillContentPermission(ctx, contentDetailsList, op)

	log.Info(ctx, "build content detail list successfully", log.Any("content", contentDetailsList))

	return contentDetailsList, nil
}

func (cm *ContentModel) listAllScopes(ctx context.Context, operator *entity.Operator) ([]string, error) {
	schools, err := external.GetSchoolServiceProvider().GetByOrganizationID(ctx, operator, operator.OrgID)
	if err != nil {
		log.Warn(ctx, "can't get schools from org", log.Err(err))
		return nil, err
	}
	ret := []string{operator.OrgID}
	for i := range schools {
		ret = append(ret, schools[i].ID)
	}

	return ret, nil
}

func (cm *ContentModel) getContentVisibilitySettings(ctx context.Context, cid string) (*entity.ContentVisibilitySettings, error) {
	contentVisibilitySettings, err := da.GetContentDA().GetContentVisibilitySettings(ctx, dbo.MustGetDB(ctx), cid)
	if err != nil {
		log.Error(ctx, "GetContentVisibilitySettings",
			log.Err(err),
			log.String("id", cid))
		return nil, err
	}
	res := &entity.ContentVisibilitySettings{
		ContentID:          cid,
		VisibilitySettings: contentVisibilitySettings,
	}
	log.Info(ctx, "ContentVisibilitySettings",
		log.Strings("visibilitySettings", contentVisibilitySettings),
		log.String("id", cid))
	return res, nil
}
func (cm *ContentModel) getContentProperties(ctx context.Context, cid string) (*entity.ContentProperties, error) {
	contentProperties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, dbo.MustGetDB(ctx), []string{cid})
	if err != nil {
		log.Error(ctx, "BatchGetByContentIDList",
			log.Err(err),
			log.String("id", cid))
		return nil, err
	}

	subjects := make([]string, 0)
	categories := make([]string, 0)
	subCategories := make([]string, 0)
	ages := make([]string, 0)
	grades := make([]string, 0)
	program := ""

	for i := range contentProperties {
		switch contentProperties[i].PropertyType {
		case entity.ContentPropertyTypeProgram:
			program = contentProperties[i].PropertyID
		case entity.ContentPropertyTypeSubject:
			subjects = append(subjects, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeCategory:
			categories = append(categories, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeAge:
			ages = append(ages, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeGrade:
			grades = append(grades, contentProperties[i].PropertyID)
		case entity.ContentPropertyTypeSubCategory:
			subCategories = append(subCategories, contentProperties[i].PropertyID)
		}
	}
	return &entity.ContentProperties{
		ContentID:   cid,
		Program:     program,
		Subject:     subjects,
		Category:    categories,
		SubCategory: subCategories,
		Age:         ages,
		Grade:       grades,
	}, nil
}

func (cm *ContentModel) buildFolderConditionWithPermission(ctx context.Context, user *entity.Operator, condition *entity.ContentConditionRequest, searchUserIDs []string, perm map[external.PermissionName]bool) *da.FolderCondition {
	dirPath := condition.DirPath
	isAssets := false
	disableFolder := true
	for i := range condition.ContentType {
		if entity.NewContentType(condition.ContentType[i]).IsAsset() {
			isAssets = true
			continue
		}
		if entity.NewContentType(condition.ContentType[i]) == entity.AliasContentTypeFolder {
			disableFolder = false
		}
	}
	partition := entity.FolderPartitionMaterialAndPlans
	if isAssets {
		partition = entity.FolderPartitionAssets
	}

	folderCondition := &da.FolderCondition{
		OwnerType:    int(entity.OwnerTypeOrganization),
		ItemType:     int(entity.FolderItemTypeFolder),
		Owner:        user.OrgID,
		NameLike:     condition.Name,
		Name:         condition.ContentName,
		ExactDirPath: dirPath,
		//Editors:      searchUserIDs,
		Partition: partition,
		Disable:   disableFolder,
	}
	if len(condition.PublishStatus) == 1 && condition.PublishStatus[0] == entity.ContentStatusPublished {
		folderCondition.ShowEmptyFolder = entity.NullBool{Bool: perm[external.ShowAllFolder295], Valid: true}
	}
	return folderCondition
}

func (cm *ContentModel) buildFolderCondition(ctx context.Context, condition *entity.ContentConditionRequest, searchUserIDs []string, user *entity.Operator) *da.FolderCondition {
	dirPath := condition.DirPath
	isAssets := false
	disableFolder := true
	for i := range condition.ContentType {
		if entity.NewContentType(condition.ContentType[i]).IsAsset() {
			isAssets = true
			continue
		}
		if entity.NewContentType(condition.ContentType[i]) == entity.AliasContentTypeFolder {
			disableFolder = false
		}
	}
	partition := entity.FolderPartitionMaterialAndPlans
	if isAssets {
		partition = entity.FolderPartitionAssets
	}

	folderCondition := &da.FolderCondition{
		OwnerType:    int(entity.OwnerTypeOrganization),
		ItemType:     int(entity.FolderItemTypeFolder),
		Owner:        user.OrgID,
		NameLike:     condition.Name,
		Name:         condition.ContentName,
		ExactDirPath: dirPath,
		//Editors:      searchUserIDs,
		Partition: partition,
		Disable:   disableFolder,
	}
	return folderCondition
}

func (cm *ContentModel) hasFolderPermission(ctx context.Context, user *entity.Operator) (bool, error) {
	return external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, user, external.CreateFolder289)
}

func (cm *ContentModel) fillFolderContentPermissionSlim(ctx context.Context, user *entity.Operator, objs []*entity.FolderContentData, perm map[external.PermissionName]bool) {
	var contentList []*entity.ContentInfo
	for i := range objs {
		//init permission
		objs[i].Permission = entity.ContentPermission{
			ID:             objs[i].ID,
			AllowEdit:      perm[external.CreateFolder289],
			AllowDelete:    perm[external.CreateFolder289],
			AllowApprove:   false,
			AllowReject:    false,
			AllowRepublish: false,
		}
		objs[i].ContentTypeName = objs[i].ContentType.Name()

		if objs[i].ContentType != entity.AliasContentTypeFolder {
			//contentIDs = append(contentIDs, objs[i].ID)
			contentInfo := &entity.ContentInfo{
				ID:            objs[i].ID,
				ContentType:   objs[i].ContentType,
				Name:          objs[i].ContentName,
				Keywords:      objs[i].Keywords,
				Description:   objs[i].Description,
				Thumbnail:     objs[i].Thumbnail,
				Data:          objs[i].Data,
				Author:        objs[i].Author,
				PublishStatus: entity.ContentPublishStatus(objs[i].PublishStatus),
				CreatedAt:     int64(objs[i].CreateAt),
				UpdatedAt:     int64(objs[i].UpdateAt),
			}
			contentList = append(contentList, contentInfo)
			//when it comes to content, set permission as false, pending to check content permission
			objs[i].Permission.AllowEdit = false
			objs[i].Permission.AllowDelete = false
		}
	}

	log.Debug(ctx, "hasFolderPermission result after init permissions",
		log.Any("contentList", contentList),
		log.Any("objs", objs))

	//build content profiles
	contentProfiles, err := cm.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		log.Error(ctx, "buildContentProfiles failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentList", contentList))
		return
	}
	log.Debug(ctx, "contentProfiles result",
		log.Any("contentList", contentList),
		log.Any("contentProfiles", contentProfiles))

	//get permission map
	permissionMap, err := GetContentPermissionChecker().BatchGetContentPermission(ctx, user, contentProfiles)
	if err != nil {
		log.Error(ctx, "BatchGetContentPermissions failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentProfiles", contentProfiles),
			log.Any("contentList", contentList))
		return
	}
	log.Debug(ctx, "BatchGetContentPermissions result",
		log.Any("permissionMap", permissionMap))
	//if permission is in the map, replace the permission of content
	for i := range objs {
		p, exists := permissionMap[objs[i].ID]
		if exists {
			objs[i].Permission = p
		}
	}
	log.Debug(ctx, "fillFolderContentPermission result",
		log.Any("objs", objs))
}
func (cm *ContentModel) fillFolderContentPermission(ctx context.Context, objs []*entity.FolderContentData, user *entity.Operator) {
	hasFolderPermission, err := cm.hasFolderPermission(ctx, user)
	if err != nil {
		log.Error(ctx, "hasFolderPermission failed",
			log.Err(err),
			log.Any("user", user))
		return
	}
	log.Debug(ctx, "hasFolderPermission result",
		log.Any("hasFolderPermission", hasFolderPermission),
		log.Any("objs", objs),
		log.Any("user", user))

	contentIDs := make([]string, len(objs))
	for i := range objs {
		//init permission
		objs[i].Permission = entity.ContentPermission{
			ID:             objs[i].ID,
			AllowEdit:      hasFolderPermission,
			AllowDelete:    hasFolderPermission,
			AllowApprove:   false,
			AllowReject:    false,
			AllowRepublish: false,
		}

		if objs[i].ContentType != entity.AliasContentTypeFolder {
			contentIDs = append(contentIDs, objs[i].ID)
			//when it comes to content, set permission as false, pending to check content permission
			objs[i].Permission.AllowEdit = false
			objs[i].Permission.AllowDelete = false
		}
	}

	log.Debug(ctx, "hasFolderPermission result after init permissions",
		log.Any("objs", objs))

	contentDetailsList, err := cm.GetContentByIDList(ctx, dbo.MustGetDB(ctx), contentIDs, user)
	if err != nil {
		log.Error(ctx, "GetContentByIDList failed",
			log.Err(err),
			log.Any("user", user),
			log.Strings("contentIDs", contentIDs))
		return
	}

	contentList := make([]*entity.ContentInfo, len(contentDetailsList))
	for i := range contentDetailsList {
		contentList[i] = &contentDetailsList[i].ContentInfo
	}
	log.Debug(ctx, "getContentInfoByIDList result",
		log.Any("contentDetailsList", contentDetailsList),
		log.Any("contentList", contentList),
		log.Any("objs", objs))

	//build content profiles
	contentProfiles, err := cm.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		log.Error(ctx, "buildContentProfiles failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentList", contentList))
		return
	}
	log.Debug(ctx, "contentProfiles result",
		log.Any("contentList", contentList),
		log.Any("contentProfiles", contentProfiles))

	//get permission map
	permissionMap, err := GetContentPermissionChecker().BatchGetContentPermission(ctx, user, contentProfiles)
	if err != nil {
		log.Error(ctx, "BatchGetContentPermissions failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentProfiles", contentProfiles),
			log.Any("contentList", contentList))
		return
	}
	log.Debug(ctx, "BatchGetContentPermissions result",
		log.Any("permissionMap", permissionMap))
	//if permission is in the map, replace the permission of content
	for i := range objs {
		p, exists := permissionMap[objs[i].ID]
		if exists {
			objs[i].Permission = p
		}
	}
	log.Debug(ctx, "fillFolderContentPermission result",
		log.Any("objs", objs))
}

func (cm *ContentModel) fillContentPermission(ctx context.Context, objs []*entity.ContentInfoWithDetails, user *entity.Operator) {
	contentList := make([]*entity.ContentInfo, len(objs))
	for i := range objs {
		contentList[i] = &objs[i].ContentInfo
		objs[i].Permission = entity.ContentPermission{
			ID:             objs[i].ID,
			AllowEdit:      false,
			AllowDelete:    false,
			AllowApprove:   false,
			AllowReject:    false,
			AllowRepublish: false,
		}
	}
	log.Debug(ctx, "fillContentPermission",
		log.Any("contentList", contentList),
		log.Any("objs", objs))

	contentProfiles, err := cm.buildContentProfiles(ctx, contentList, user)
	if err != nil {
		log.Error(ctx, "buildContentProfiles failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentList", contentList))
		return
	}

	log.Debug(ctx, "buildContentProfiles result",
		log.Any("contentProfiles", contentProfiles))

	permissionMap, err := GetContentPermissionChecker().BatchGetContentPermission(ctx, user, contentProfiles)
	if err != nil {
		log.Error(ctx, "BatchGetContentPermissions failed",
			log.Err(err),
			log.Any("user", user),
			log.Any("contentProfiles", contentProfiles),
			log.Any("contentList", contentList))
		return
	}

	log.Debug(ctx, "BatchGetContentPermissions result",
		log.Any("permissionMap", permissionMap))

	for i := range objs {
		p, exists := permissionMap[objs[i].ID]
		if exists {
			objs[i].Permission = p
		}
	}
	log.Debug(ctx, "fillContentPermission result",
		log.Any("objs", objs),
		log.Any("permissionMap", permissionMap))
}

func (c *ContentModel) buildContentProfiles(ctx context.Context, content []*entity.ContentInfo, user *entity.Operator) ([]*ContentEntityProfile, error) {
	profiles := make([]*ContentEntityProfile, len(content))

	schoolsInfo, err := GetContentFilterModel().QueryUserSchools(ctx, user)
	if err != nil {
		log.Error(ctx, "getVisibilitySettingsType failed",
			log.Err(err),
			log.Any("content", content),
			log.Any("user", user))
		return nil, err
	}
	log.Debug(ctx, "querySchools result",
		log.Any("schoolsInfo", schoolsInfo),
		log.Any("content", content))

	for i := range content {
		visibilitySettingType, err := c.getVisibilitySettingsType(ctx, content[i].PublishScope, schoolsInfo, user)
		if err != nil {
			log.Error(ctx, "getVisibilitySettingsType failed",
				log.Err(err),
				log.Any("content", content))
			return nil, err
		}

		ownerType := OwnerTypeUser
		if user.UserID != content[i].Author {
			ownerType = OwnerTypeOthers
		}
		profiles[i] = &ContentEntityProfile{
			ID: content[i].ID,
			ContentProfile: ContentProfile{
				ContentType:        content[i].ContentType,
				Status:             content[i].PublishStatus,
				VisibilitySettings: visibilitySettingType,
				Owner:              ownerType,
			},
		}
	}
	return profiles, nil
}

func (c *ContentModel) getVisibilitySettingsType(ctx context.Context, visibilitySettings []string, schoolInfo *contentFilterUserSchoolInfo, user *entity.Operator) (VisibilitySettingsType, error) {
	containsOrg := false
	containsOtherSchools := false
	containsSchools := false
	for i := range visibilitySettings {
		if visibilitySettings[i] == user.OrgID {
			//contains org
			containsOrg = true
		} else {
			containsSchools = true
			if !utils.CheckInStringArray(visibilitySettings[i], schoolInfo.MySchool) {
				if utils.CheckInStringArray(visibilitySettings[i], schoolInfo.AllSchool) {
					//contains other schools in org
					containsOtherSchools = true
				} else {
					log.Warn(ctx, "visibility setting is not in all schools",
						log.Strings("visibilitySettings", visibilitySettings),
						log.Any("mySchool", schoolInfo.MySchool),
						log.Any("allSchool", schoolInfo.AllSchool),
						log.Any("user", user))
					return VisibilitySettingsTypeAllSchools, ErrInvalidVisibilitySetting
				}
			}
		}
	}
	log.Info(ctx, "visibility settings check result",
		log.Strings("visibilitySettings", visibilitySettings),
		log.Bool("containsOrg", containsOrg),
		log.Bool("containsOtherSchools", containsOtherSchools),
		log.Bool("containsSchools", containsSchools))

	//contains org
	if containsOrg {
		//contains other schools
		if containsOtherSchools {
			return VisibilitySettingsTypeOrgWithAllSchools, nil
		}
		if !containsSchools {
			//only contains org
			return VisibilitySettingsTypeOnlyOrg, nil
		}
		return VisibilitySettingsTypeOrgWithMySchools, nil
	}

	//contains other schools but org
	if containsOtherSchools {
		return VisibilitySettingsTypeAllSchools, nil
	}
	//only contains my schools
	return VisibilitySettingsTypeMySchools, nil
}

func (cm *ContentModel) fillFolderContent(ctx context.Context, objs []*entity.FolderContentData, user *entity.Operator) {
	authorIDs := make([]string, len(objs))
	for i := range objs {
		authorIDs[i] = objs[i].Author
	}

	authorMap, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, user, authorIDs)
	if err != nil {
		log.Warn(ctx, "get user info failed", log.Err(err), log.Any("objs", objs))
	}

	for i := range objs {
		objs[i].AuthorName = authorMap[objs[i].Author]
		objs[i].ContentTypeName = objs[i].ContentType.Name()
	}

	//fill folder content permissions
	cm.fillFolderContentPermission(ctx, objs, user)
}

func (cm *ContentModel) conditionRequestToCondition(c entity.ContentConditionRequest) *da.ContentCondition {
	return &da.ContentCondition{
		Name:               c.Name,
		ContentType:        c.ContentType,
		VisibilitySettings: c.VisibilitySettings,
		PublishStatus:      c.PublishStatus,
		Author:             c.Author,
		Org:                c.Org,
		Program:            c.Program,
		SourceType:         c.SourceType,
		DirPath: entity.NullStrings{
			Strings: []string{c.DirPath},
			Valid:   c.DirPath != "",
		},
		ContentName:    c.ContentName,
		IDS:            c.ContentIDs,
		ParentPath:     c.ParentsPath,
		OrderBy:        da.NewContentOrderBy(c.OrderBy),
		Pager:          c.Pager,
		JoinUserIDList: c.JoinUserIDList,

		UseJoinForVisibilitySettings: c.UseJoinForVisibilitySettings,
	}
}

func (cm *ContentModel) convertFolderContent(ctx context.Context, objs []*entity.FolderContent, user *entity.Operator) []*entity.FolderContentData {
	ret := make([]*entity.FolderContentData, len(objs))
	for i := range objs {
		ret[i] = &entity.FolderContentData{
			ID:              objs[i].ID,
			ContentName:     objs[i].ContentName,
			ContentType:     objs[i].ContentType,
			Description:     objs[i].Description,
			Keywords:        strings.Split(objs[i].Keywords, constant.StringArraySeparator),
			Author:          objs[i].Author,
			ItemsCount:      objs[i].ItemsCount,
			PublishStatus:   objs[i].PublishStatus,
			Thumbnail:       objs[i].Thumbnail,
			Data:            objs[i].Data,
			AuthorName:      objs[i].Author,
			DirPath:         objs[i].DirPath,
			ContentTypeName: objs[i].ContentTypeName,
			CreateAt:        objs[i].CreateAt,
			UpdateAt:        objs[i].UpdateAt,
		}
	}
	return ret
}

func (cm *ContentModel) GetContentByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string) ([]*entity.ContentInfoInternal, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make([]*entity.ContentInfoInternal, 0)

	nid, cachedContent := da.GetContentRedis().GetContentCacheByIDList(ctx, cids)
	for i := range cachedContent {
		fileType, err := cm.parseFileType(ctx, cachedContent[i].ContentType, cachedContent[i].Data)
		if err != nil {
			log.Error(ctx, "parse file type error", log.Err(err), log.Any("content", cachedContent[i]))
			return nil, err
		}
		resp = append(resp, &entity.ContentInfoInternal{
			ID:          cachedContent[i].ID,
			Name:        cachedContent[i].Name,
			ContentType: cachedContent[i].ContentType,
			LatestID:    cachedContent[i].LatestID,
			OutcomeIDs:  cachedContent[i].Outcomes,
			FileType:    fileType,
		})
	}
	if len(nid) < 1 {
		return resp, nil
	}

	data, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IncludeDeleted: true,
		IDS: entity.NullStrings{
			Strings: nid,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		var latestID = data[i].LatestID
		if data[i].LatestID == "" {
			latestID = data[i].ID
		}
		fileType, err := cm.parseFileType(ctx, data[i].ContentType, data[i].Data)
		if err != nil {
			log.Error(ctx, "parse file type error", log.Err(err), log.Any("content", data[i]))
			return nil, err
		}
		resp = append(resp, &entity.ContentInfoInternal{
			ID:          data[i].ID,
			Name:        data[i].Name,
			ContentType: data[i].ContentType,
			LatestID:    latestID,
			OutcomeIDs:  cm.parseContentOutcomes(ctx, data[i]),
			FileType:    fileType,
		})
	}
	return resp, nil
}

func (cm *ContentModel) parseFileType(ctx context.Context, contentType entity.ContentType, dataStr string) (entity.FileType, error) {
	data, err := cm.CreateContentData(ctx, contentType, dataStr)
	if err != nil {
		log.Error(ctx, "get lesson material source map: create content data failed",
			log.Err(err),
			log.Any("contentType", contentType),
			log.String("dataStr", dataStr),
		)
		return 0, err
	}

	if v, ok := data.(*MaterialData); ok {
		return v.FileType, nil
	}

	return 0, nil
}

func (cm *ContentModel) GetContentsSubContentsMapByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string, user *entity.Operator) (map[string][]*entity.ContentInfoInternal, error) {
	objs, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IncludeDeleted: true,
		IDS: entity.NullStrings{
			Strings: cids,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "can't read content", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	contentInfoMap := make(map[string][]*entity.ContentInfoInternal)
	for _, obj := range objs {
		cd, err := cm.CreateContentData(ctx, obj.ContentType, obj.Data)
		if err != nil {
			log.Error(ctx, "can't unmarshal contentdata", log.Err(err), log.Any("content", obj))
			return nil, err
		}

		switch v := cd.(type) {
		case *LessonData:
			//存在子内容，则返回子内容
			//if the content contains sub contents, return sub contents
			content, err := cm.ConvertContentObj(ctx, tx, obj, user)
			if err != nil {
				log.Error(ctx, "can't parse contentdata", log.Err(err))
				return nil, ErrParseContentDataFailed
			}
			err = v.PrepareVersion(ctx)
			if err != nil {
				log.Error(ctx, "can't prepare version for sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			err = v.PrepareResult(ctx, tx, content, user, true)
			if err != nil {
				log.Error(ctx, "can't get sub contents", log.Err(err), log.Any("content", content))
				return nil, err
			}
			ret := make([]*entity.ContentInfoInternal, 0)
			v.lessonDataIteratorLoop(ctx, func(ctx context.Context, l *LessonData) {
				if l.Material != nil {
					fileType, err := cm.parseFileType(ctx, l.Material.ContentType, l.Material.Data)
					if err != nil {
						return
					}
					ret = append(ret, &entity.ContentInfoInternal{
						ID:          l.Material.ID,
						Name:        l.Material.Name,
						ContentType: l.Material.ContentType,
						OutcomeIDs:  l.Material.Outcomes,
						LatestID:    l.Material.ID,
						FileType:    fileType,
					})
				}
			})
			contentInfoMap[obj.ID] = ret
		case *MaterialData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*entity.ContentInfoInternal{
				{
					ID:          obj.ID,
					Name:        obj.Name,
					ContentType: obj.ContentType,
					OutcomeIDs:  cm.parseContentOutcomes(ctx, obj),
					LatestID:    obj.ID,
					FileType:    v.FileType,
				},
			}
			contentInfoMap[obj.ID] = ret
		case *AssetsData:
			//若不存在子内容，则返回当前内容
			//if sub contents is not exists, return current content
			ret := []*entity.ContentInfoInternal{
				{
					ID:          obj.ID,
					Name:        obj.Name,
					ContentType: obj.ContentType,
					OutcomeIDs:  cm.parseContentOutcomes(ctx, obj),
					LatestID:    obj.ID,
					FileType:    v.FileType,
				},
			}
			contentInfoMap[obj.ID] = ret
		}
	}

	return contentInfoMap, nil
}

func (cm *ContentModel) GetLatestContentIDMapByIDListInternal(ctx context.Context, tx *dbo.DBContext, cids []string) (map[string]string, error) {
	if len(cids) < 1 {
		return nil, nil
	}
	resp := make(map[string]string)
	data, err := da.GetContentDA().QueryContent(ctx, tx, &da.ContentCondition{
		IncludeDeleted: true,
		IDS: entity.NullStrings{
			Strings: cids,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, ErrReadContentFailed
	}
	for i := range data {
		latestID := data[i].LatestID
		if data[i].LatestID == "" {
			latestID = data[i].ID
		}
		resp[data[i].ID] = latestID
	}

	return resp, nil
}

func (cm *ContentModel) CleanCache(ctx context.Context) {
	da.GetContentDA().CleanCache(ctx)
}

var (
	_contentModel     IContentModel
	_contentModelOnce sync.Once
)

func GetContentModel() IContentModel {
	_contentModelOnce.Do(func() {
		m := new(ContentModel)

		sharedContentCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  da.RedisKeyPrefixContentShared,
			Expiration:      constant.SharedContentQueryCacheExpiration,
			RefreshDuration: constant.SharedContentQueryCacheRefreshDuration,
			RawQuery:        m.searchSharedContent})
		if err != nil {
			log.Panic(context.Background(), "create shared content query cache failed", log.Err(err))
		}

		m.sharedContentQueryCache = sharedContentCache

		_contentModel = m
	})

	return _contentModel
}
