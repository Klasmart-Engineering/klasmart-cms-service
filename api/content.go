package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/gin-gonic/gin"
)

type contentBulkOperateRequest struct {
	ID []string `json:"id"`
}

type CreateContentResponse struct {
	ID string `json:"id"`
}

type CreateFolderResponse struct {
	ID string `json:"id"`
}
type PublishContentRequest struct {
	Scope []string `json:"scope"`
}

// @Summary createContent
// @ID createContent
// @Description create lesson plan, lesson material or assets
// @Accept json
// @Produce json
// @Param content body entity.CreateContentRequest true "create request"
// @Tags content
// @Success 200 {object} CreateContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents [post]
func (s *Server) createContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	if data.ContentType == entity.ContentTypeAssets {
		data.PublishScope = []string{op.OrgID}
	}

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckCreateContentPermission(ctx, data, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	cid, err := model.GetContentModel().CreateContent(ctx, data, op)

	switch err {
	case model.ErrContentDataRequestSource:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidSourceExt:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrTeacherManualExtension:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
	case model.ErrInvalidParentFolderId:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrSuggestTimeTooSmall:
		c.JSON(http.StatusBadRequest, L(LibraryErrorPlanDuration))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidSelectForm:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary publishContentBulk
// @ID publishContentBulk
// @Description publish contents bulk
// @Accept json
// @Produce json
// @Param contentIds body contentBulkOperateRequest true "content bulk id list"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_bulk/publish [put]
func (s *Server) publishContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckRepublishContentsPermission(ctx, ids.ID, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	err = model.GetContentModel().PublishContentBulkTx(ctx, ids.ID, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, L(GeneralUnknown))
	case model.ErrInvalidContentStatusToPublish:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrPlanHasArchivedMaterials:
		c.JSON(http.StatusBadRequest, L(LibraryIncludeArchivedMaterials))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary publishContent
// @ID publishContent
// @Description publish a content
// @Accept json
// @Produce json
// @Param content_id path string true "content id to publish"
// @Param data body PublishContentRequest true "content publish data"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id}/publish [put]
func (s *Server) publishContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")

	data := new(PublishContentRequest)
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckPublishContentsPermission(ctx, cid, data.Scope, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	err = model.GetContentModel().PublishContentTx(ctx, cid, data.Scope, op)

	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrPlanHasArchivedMaterials:
		c.JSON(http.StatusBadRequest, L(LibraryIncludeArchivedMaterials))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary publishContentWithAssets
// @ID publishContentWithAssets
// @Description publish a content with assets
// @Accept json
// @Produce json
// @Param content_id path string true "content id to publish"
// @Param data body PublishContentRequest true "content publish data"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id}/publish/assets [put]
func (s *Server) publishContentWithAssets(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")

	data := new(PublishContentRequest)
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckPublishContentsPermission(ctx, cid, data.Scope, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}
	err = model.GetContentModel().PublishContentWithAssetsTx(ctx, cid, data.Scope, op)

	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrPlanHasArchivedMaterials:
		c.JSON(http.StatusBadRequest, L(LibraryIncludeArchivedMaterials))
	case model.ErrInvalidSourceExt:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrTeacherManualExtension:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary getContent
// @ID getContentById
// @Description get a content by id (Inherent & unchangeable)
// @Accept json
// @Produce json
// @Param content_id path string true "get content id"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetails
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id} [get]
func (s *Server) getContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")
	var data struct {
		Scope string `json:"scope"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckGetContentPermission(ctx, cid, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	result, err := model.GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary updateContent
// @ID updateContent
// @Description update a content data
// @Accept json
// @Produce json
// @Param content_id path string true "content id to publish"
// @Param contentData body entity.CreateContentRequest true "content data to update"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id} [put]
func (s *Server) updateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckUpdateContentPermission(ctx, cid, op)
	if err != nil {
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
			return
		}
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	err = model.GetContentModel().UpdateContent(ctx, dbo.MustGetDB(ctx), cid, data, op)
	switch err {
	case model.ErrContentDataRequestSource:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrSuggestTimeTooSmall:
		c.JSON(http.StatusBadRequest, L(LibraryErrorPlanDuration))
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidSourceExt:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrTeacherManualExtension:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary lockContent
// @ID lockContent
// @Description lock a content to edit
// @Accept json
// @Produce json
// @Param content_id path string true "content id to lock"
// @Tags content
// @Success 200 {object} CreateContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id}/lock [put]
func (s *Server) lockContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckUpdateContentPermission(ctx, cid, op)
	if err != nil {
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
			return
		}
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	ncid, err := model.GetContentModel().LockContentTx(ctx, cid, op)
	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidLockedContentPublishStatus:
		c.JSON(http.StatusConflict, L(LibraryContentLockedByMe))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": ncid,
		})
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary deleteContentBulk
// @ID deleteContentBulk
// @Description delete contents bulk
// @Accept json
// @Produce json
// @Param contentIds body contentBulkOperateRequest true "content bulk id list"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_bulk [delete]
func (s *Server) deleteContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckDeleteContentPermission(ctx, ids.ID, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	err = model.GetContentModel().DeleteContentBulkTx(ctx, ids.ID, op)

	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary deleteContent
// @ID deleteContent
// @Description delete a content
// @Accept json
// @Produce json
// @Param content_id path string true "content id to delete"
// @Tags content
// @Success 200 {object} string ok
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id} [delete]
func (s *Server) deleteContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckDeleteContentPermission(ctx, []string{cid}, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	err = model.GetContentModel().DeleteContentTx(ctx, cid, op)

	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary contentDataCount
// @ID getContentsStatistics
// @Description get content data count
// @Accept json
// @Produce json
// @Param content_id path string true "content id to get count"
// @Tags content
// @Success 200 {object} entity.ContentStatisticsInfo
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/{content_id}/statistics [get]
func (s *Server) contentDataCount(c *gin.Context) {
	ctx := c.Request.Context()
	cid := c.Param("content_id")
	res, err := model.GetContentModel().ContentDataCount(ctx, dbo.MustGetDB(ctx), cid)
	switch err {
	case nil:
		c.JSON(http.StatusOK, res)
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary queryContent
// @ID searchContents
// @Description query content by condition (Inherent & unchangeable)
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param author query string false "search content author"
// @Param content_type query string false "search content type"
// @Param scope query string false "search content scope"
// @Param program_group query string false "search program group"
// @Param submenu query string false "search page sub menu"
// @Param program query string false "search content program"
// @Param content_name query string false "search content name"
// @Param path query string false "search content path"
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.QueryContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents [get]
func (s *Server) queryContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.queryContentCondition(c, op)
	author := c.Query("author")
	//filter unauthed visibility settings
	if author != constant.Self {
		isTerminal := s.filterPublishedContent(c, &condition, op)
		if isTerminal {
			return
		}
	}

	if condition.PublishedQueryMode != entity.PublishedQueryModeOnlyOwner {
		hasPermission, err := model.GetContentPermissionMySchoolModel().CheckQueryContentPermission(ctx, &condition, op)
		if err != nil {
			s.defaultErrorHandler(c, err)
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, L(GeneralNoPermission))
			return
		}
	}

	total := 0
	var err error
	var results []*entity.ContentInfoWithDetails
	if author == constant.Self {
		total, results, err = model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	} else {
		total, results, err = model.GetContentModel().SearchUserContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, s.convertQueryContentResult(ctx, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		}))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary querySharedContent
// @ID querySharedContent
// @Description query authed content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param content_type query string false "search content type"
// @Param program query string false "search content program"
// @Param content_name query string false "search content name"
// @Param program_group query string false "search program group"
// @Param submenu query string false "search page sub menu"
// @Param source_type query string false "search content source type"
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.QueryContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_authed [get]
func (s *Server) querySharedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.queryContentCondition(c, op)
	total, results, err := model.GetContentModel().SearchSharedContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s.convertQueryContentResult(ctx, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		}))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary querySharedContentV2
// @ID querySharedContentV2
// @Description query shared content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param content_type query string false "search content type"
// @Param program query string false "search content program"
// @Param content_name query string false "search content name"
// @Param program_group query string false "search program group"
// @Param submenu query string false "search page sub menu"
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Param parent_id query string false "parent_id"
// @Tags content
// @Success 200 {object} entity.QuerySharedContentV2Response
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_shared [get]
func (s *Server) querySharedContentV2(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.querySharedContentCondition(c, op)
	condition.ParentID = c.Query("parent_id")
	if condition.ParentID == "" {
		condition.ParentID = "/"
	}
	response, err := model.GetContentModel().SearchSharedContentV2(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, response)
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

func (s *Server) convertQueryContentResult(ctx context.Context, cdr *entity.ContentInfoWithDetailsResponse) (qcr *entity.QueryContentResponse) {
	qcr = &entity.QueryContentResponse{
		List: []*entity.QueryContentItem{},
	}
	if cdr == nil {
		return
	}
	qcr.Total = cdr.Total
	for _, cd := range cdr.ContentList {
		qcr.List = append(qcr.List, &entity.QueryContentItem{
			ID:              cd.ID,
			ContentType:     cd.ContentType,
			Name:            cd.Name,
			Thumbnail:       cd.Thumbnail,
			AuthorName:      cd.AuthorName,
			Data:            cd.Data,
			Author:          cd.Author,
			PublishStatus:   cd.PublishStatus,
			ContentTypeName: cd.ContentTypeName,
			Permission:      cd.Permission,
			SuggestTime:     cd.SuggestTime,
		})
	}
	return
}

// @Summary queryFolderContent
// @ID queryFolderContent
// @Description query content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param author query string false "search content author"
// @Param content_type query string false "search content type"
// @Param scope query string false "search content scope"
// @Param content_name query string false "search content name"
// @Param submenu query string false "search page sub menu"
// @Param program query string false "search content program"
// @Param program_group query string false "search program group"
// @Param path query string false "search content path"
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.FolderContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} BadRequestResponse
// @Router /contents_folders [get]
func (s *Server) queryFolderContent(c *gin.Context) {
	utils.InsertCacheIntoGinCtx(c)
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.queryContentCondition(c, op)
	author := c.Query("author")

	//if query is not self, filter conditions
	if author != constant.Self {
		isTerminal := s.filterPublishedContent(c, &condition, op)
		if isTerminal {
			return
		}
	}
	if condition.PublishedQueryMode != entity.PublishedQueryModeOnlyOwner {
		hasPermission, err := model.GetContentPermissionMySchoolModel().CheckQueryContentPermission(ctx, &condition, op)
		if err != nil {
			s.defaultErrorHandler(c, err)
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, L(GeneralNoPermission))
			return
		}
	}

	total := 0
	var err error
	var results []*entity.FolderContentData
	if author == constant.Self {
		total, results, err = model.GetContentModel().SearchUserPrivateFolderContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	} else {
		total, results, err = model.GetContentModel().SearchUserFolderContentSlim(ctx, dbo.MustGetDB(ctx), &condition, op)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.FolderContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	case model.ErrInvalidVisibleScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary queryPrivateContent
// @ID searchPrivateContents
// @Description query private content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param author query string false "search content author"
// @Param content_type query string false "search content type"
// @Param program query string false "search content program"
// @Param program_group query string false "search program group"
// @Param content_name query string false "search content name"
// @Param submenu query string false "search page sub menu"
// @Param source_type query string false "search content source type"
// @Param scope query string false "search content scope"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.QueryContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} BadRequestResponse
// @Router /contents_private [get]
func (s *Server) queryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	condition := s.queryContentCondition(c, op)
	condition.Author = op.UserID

	hasPermission, err := model.GetContentPermissionMySchoolModel().CheckQueryContentPermission(ctx, &condition, op)
	if err != nil {
		s.defaultErrorHandler(c, err)
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
		return
	}

	total, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s.convertQueryContentResult(ctx, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		}))
	case model.ErrInvalidVisibleScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary queryPendingContent
// @ID searchPendingContents
// @Description query pending content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param author query string false "search content author"
// @Param content_type query string false "search content type"
// @Param scope query string false "search content scope"
// @Param program query string false "search content program"
// @Param submenu query string false "search page sub menu"
// @Param content_name query string false "search content name"
// @Param program_group query string false "search program group"
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.QueryContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_pending [get]
func (s *Server) queryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op := s.getOperator(c)

	condition := s.queryContentCondition(c, op)
	//filter pending visibility settings
	isTerminal := s.filterPendingContent(c, &condition, op)
	if isTerminal {
		return
	}

	if condition.PublishedQueryMode != entity.PublishedQueryModeOnlyOwner {
		hasPermission, err := model.GetContentPermissionMySchoolModel().CheckQueryContentPermission(ctx, &condition, op)
		if err != nil {
			s.defaultErrorHandler(c, err)
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, L(GeneralNoPermission))
			return
		}
	}

	total, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, s.convertQueryContentResult(ctx, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		}))
	case model.ErrInvalidVisibilitySetting:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoPermissionToQuery:
		c.JSON(http.StatusForbidden, L(GeneralNoPermission))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary queryContentInternal
// @ID queryContentInternal
// @Description query content internal
// @Produce json
// @Param org_id query string false "search content under the organization"
// @Param content_ids query string false "search content id list, separated by commas"
// @Param content_type query int false "search content type, 1 for materials & 2 for plans"
// @Param create_at_le query int false "search content create_at less"
// @Param create_at_ge query int false "search content create_at greater"
// @Param plan_id query string false "search materials from lesson_plan"
// @Param schedule_id query string false "search student content map under the schedule review"
// @Param source_id query string false "search content by source id"
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentSimplifiedList
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /internal/contents [get]
func (s *Server) queryContentInternal(c *gin.Context) {
	ctx := c.Request.Context()
	condition := s.queryContentInternalCondition(c)

	result, err := model.GetContentModel().SearchSimplifyContentInternal(ctx, dbo.MustGetDB(ctx), &condition)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

// @Summary getSTMLessonPlan
// @ID getSTMLessonPlan
// @Description get stm lesson_plan
// @Produce json
// @Tags content
// @Param ids body []string true "stm lesson_plan"
// @Success 200 {array} entity.LessonPlan
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /internal/stm/contents [post]
func (s *Server) getSTMLessonPlan(c *gin.Context) {
	ctx := c.Request.Context()

	var data []string
	err := c.ShouldBindJSON(&data)
	if err != nil {
		log.Warn(ctx, "internal query", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	result, err := model.GetContentModel().GetSTMLessonPlans(ctx, dbo.MustGetDB(ctx), data)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	case constant.ErrInvalidArgs:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		s.defaultErrorHandler(c, err)
	}
}

func parseAuthor(c *gin.Context, u *entity.Operator) string {
	author := c.Query("author")
	if author == constant.Self {
		author = u.UserID
	}
	return author
}

func parseOrg(c *gin.Context, u *entity.Operator) string {
	author := c.Query("org")
	if author == constant.Self {
		author = u.OrgID
	}
	return author
}

func (s *Server) filterPendingContent(c *gin.Context, condition *entity.ContentConditionRequest, op *entity.Operator) bool {
	ctx := c.Request.Context()
	err := model.GetContentFilterModel().FilterPendingContent(ctx, condition, op)
	if err == model.ErrNoAvailableVisibilitySettings {
		//no available visibility settings
		c.JSON(http.StatusOK, &entity.FolderContentInfoWithDetailsResponse{
			Total:       0,
			ContentList: nil,
		})
		return true
	}
	if err != nil {
		s.defaultErrorHandler(c, err)
		return true
	}
	return false
}

func (s *Server) filterPublishedContent(c *gin.Context, condition *entity.ContentConditionRequest, op *entity.Operator) bool {
	ctx := c.Request.Context()
	var err error
	if c.Query("submenu") == constant.SubMenuLibraryArchived {
		err = model.GetContentFilterModel().FilterArchivedContent(ctx, condition, op)
	} else {
		err = model.GetContentFilterModel().FilterPublishContent(ctx, condition, op)
	}
	//no available content visibility settings, return nil
	if err == model.ErrNoAvailableVisibilitySettings {
		//no available visibility settings
		c.JSON(http.StatusOK, &entity.FolderContentInfoWithDetailsResponse{
			Total:       0,
			ContentList: make([]*entity.FolderContentData, 0),
		})
		return true
	}
	if err != nil {
		s.defaultErrorHandler(c, err)
		return true
	}
	return false
}

func (s *Server) queryContentInternalCondition(c *gin.Context) entity.ContentInternalConditionRequest {
	contentIDsStr := c.Query("content_ids")
	var contentIDs []string
	if contentIDsStr != "" {
		contentIDs = strings.Split(strings.TrimSpace(contentIDsStr), constant.StringArraySeparator)
	}
	contentType, _ := strconv.Atoi(c.Query("content_type"))
	createAtLe, _ := strconv.Atoi(c.Query("create_at_le"))
	createAtGe, _ := strconv.Atoi(c.Query("create_at_ge"))

	return entity.ContentInternalConditionRequest{
		IDs:          contentIDs,
		OrgID:        c.Query("org_id"),
		ContentType:  contentType,
		CreateAtLe:   createAtLe,
		CreateAtGe:   createAtGe,
		PlanID:       c.Query("plan_id"),
		DataSourceID: c.Query("source_id"),
		ScheduleID:   c.Query("schedule_id"),
	}
}

func (s *Server) queryContentCondition(c *gin.Context, op *entity.Operator) entity.ContentConditionRequest {
	contentTypeStr := c.Query("content_type")
	//keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	scope := c.Query("scope")
	publish := c.Query("publish_status")
	programs := c.Query("programs")
	path := c.Query("path")
	programGroup := c.Query("program_group")
	condition := entity.ContentConditionRequest{
		Author:      parseAuthor(c, op),
		Org:         parseOrg(c, op),
		DirPath:     path,
		OrderBy:     c.Query("order_by"),
		Pager:       utils.GetPager(c.Query("page"), c.Query("page_size")),
		Name:        strings.TrimSpace(c.Query("name")),
		ContentName: strings.TrimSpace(c.Query("content_name")),
	}
	sourceType := c.Query("source_type")
	//if len(keywords) > 0 {
	//	condition.Name = keywords
	//}
	if contentTypeStr != "" {
		contentTypeList := strings.Split(contentTypeStr, constant.StringArraySeparator)
		for i := range contentTypeList {
			contentType, err := strconv.Atoi(contentTypeList[i])
			if err != nil {
				log.Warn(c.Request.Context(), "parse contentType failed", log.Err(err), log.String("contentType", contentTypeStr))
				continue
			}
			ct := entity.NewContentType(contentType)
			condition.ContentType = append(condition.ContentType, ct.ContentTypeInt()...)
		}
	}
	if scope != "" {
		scopes := strings.Split(scope, constant.StringArraySeparator)
		condition.VisibilitySettings = append(condition.VisibilitySettings, scopes...)
	}
	if publish != "" {
		condition.PublishStatus = append(condition.PublishStatus, publish)
	}
	if programs != "" {
		program := strings.Split(programs, constant.StringArraySeparator)
		condition.Program = program
	}
	if programGroup != "" {
		programs, err := model.GetProgramModel().GetByGroup(c.Request.Context(), op, programGroup)
		if err != nil {
			log.Error(c.Request.Context(), "get program by groups failed", log.Err(err),
				log.String("group", programGroup))
		} else if len(programs) > 0 {
			programIDs := make([]string, len(programs))
			for i := range programs {
				programIDs[i] = programs[i].ID
			}
			condition.Program = append(condition.Program, programIDs...)
		}
	}
	if sourceType != "" {
		condition.SourceType = sourceType
	}
	return condition
}

func (s *Server) querySharedContentCondition(c *gin.Context, op *entity.Operator) entity.ContentConditionRequest {
	contentTypeStr := c.Query("content_type")
	scope := c.Query("scope")
	publish := c.Query("publish_status")
	programs := c.Query("programs")
	path := c.Query("path")
	condition := entity.ContentConditionRequest{
		Author:      parseAuthor(c, op),
		Org:         parseOrg(c, op),
		DirPath:     path,
		OrderBy:     c.Query("order_by"),
		Pager:       utils.GetPager(c.Query("page"), c.Query("page_size")),
		Name:        strings.TrimSpace(c.Query("name")),
		ContentName: strings.TrimSpace(c.Query("content_name")),
	}
	sourceType := c.Query("source_type")
	//if len(keywords) > 0 {
	//	condition.Name = keywords
	//}
	if contentTypeStr != "" {
		contentTypeList := strings.Split(contentTypeStr, constant.StringArraySeparator)
		for i := range contentTypeList {
			contentType, err := strconv.Atoi(contentTypeList[i])
			if err != nil {
				log.Warn(c.Request.Context(), "parse contentType failed", log.Err(err), log.String("contentType", contentTypeStr))
				continue
			}
			ct := entity.NewContentType(contentType)
			condition.ContentType = append(condition.ContentType, ct.ContentTypeInt()...)
		}
	}
	if scope != "" {
		scopes := strings.Split(scope, constant.StringArraySeparator)
		condition.VisibilitySettings = append(condition.VisibilitySettings, scopes...)
	}
	if publish != "" {
		condition.PublishStatus = append(condition.PublishStatus, publish)
	}
	if programs != "" {
		program := strings.Split(programs, constant.StringArraySeparator)
		condition.Program = program
	}
	if sourceType != "" {
		condition.SourceType = sourceType
	}
	return condition
}
