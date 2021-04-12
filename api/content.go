package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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

	hasPermission, err := model.GetContentPermissionModel().CheckCreateContentPermission(ctx, data, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	cid, err := model.GetContentModel().CreateContentTx(ctx, data, op)

	switch err {
	case model.ErrContentDataRequestSource:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(LibraryMsgContentDataInvalid))
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary copyContent
// @ID copyContent
// @Description copy lesson plan, lesson material
// @Accept json
// @Produce json
// @Param content body entity.CreateContentRequest true "create request"
// @Tags content
// @Success 200 {object} CreateContentResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/copy [post]
func (s *Server) copyContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.CopyContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	// hasPermission, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, op, permission)
	// if err != nil {
	// 	log.Error(ctx, "get permission failed", log.Err(err))
	// 	c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	// 	return
	// }
	// //有permission，直接返回
	// //if user has no permission return
	// if !hasPermission {
	// 	c.JSON(http.StatusForbidden, L(GeneralUnknown))
	// 	return
	// }
	cid, err := model.GetContentModel().CopyContentTx(ctx, data.ContentID, data.Deep, op)
	switch err {
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	hasPermission, err := model.GetContentPermissionModel().CheckPublishContentsPermission(ctx, ids.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	err = model.GetContentModel().PublishContentBulkTx(ctx, ids.ID, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, L(GeneralUnknown))
	case model.ErrInvalidContentStatusToPublish:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	hasPermission, err := model.GetContentPermissionModel().CheckPublishContentsPermissionBatch(ctx, cid, data.Scope, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	err = model.GetContentModel().PublishContentTx(ctx, cid, data.Scope, op)

	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	hasPermission, err := model.GetContentPermissionModel().CheckPublishContentsPermissionBatch(ctx, cid, data.Scope, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	err = model.GetContentModel().PublishContentWithAssetsTx(ctx, cid, data.Scope, op)

	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	hasPermission, err := model.GetContentPermissionModel().CheckGetContentPermission(ctx, cid, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	result, err := model.GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	hasPermission, err := model.GetContentPermissionModel().CheckUpdateContentPermission(ctx, cid, op)
	if err != nil {
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
			return
		}
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
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
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	hasPermission, err := model.GetContentPermissionModel().CheckLockContentPermission(ctx, cid, op)
	if err != nil {
		lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
		if ok {
			c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
			return
		}
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
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
	case model.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidLockedContentPublishStatus:
		c.JSON(http.StatusConflict, L(LibraryContentLockedByMe))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": ncid,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	hasPermission, err := model.GetContentPermissionModel().CheckDeleteContentPermission(ctx, ids.ID, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	err = model.GetContentModel().DeleteContentBulkTx(ctx, ids.ID, op)

	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case model.ErrDeleteLessonInSchedule:
		c.JSON(http.StatusConflict, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

	hasPermission, err := model.GetContentPermissionModel().CheckDeleteContentPermission(ctx, []string{cid}, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	err = model.GetContentModel().DeleteContentTx(ctx, cid, op)

	lockedByErr, ok := err.(*model.ErrContentAlreadyLocked)
	if ok {
		c.JSON(http.StatusNotAcceptable, LD(LibraryMsgContentLocked, lockedByErr.LockedBy))
		return
	}
	switch err {
	case model.ErrDeleteLessonInSchedule:
		c.JSON(http.StatusConflict, L(GeneralUnknown))
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
// @Param program query string false "search content program"
// @Param content_name query string false "search content name"
// @Param path query string false "search content path"
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents [get]
func (s *Server) queryContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := queryCondition(c, op)

	hasPermission, err := model.GetContentPermissionModel().CheckQueryContentPermission(ctx, condition, model.QueryModePublished, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	author := c.Query("author")
	total := 0
	var results []*entity.ContentInfoWithDetails
	if author == constant.Self {
		total, results, err = model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
	} else {
		total, results, err = model.GetContentModel().SearchUserContent(ctx, dbo.MustGetDB(ctx), condition, op)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

// @Summary queryAuthContent
// @ID queryAuthContent
// @Description query authed content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param content_type query string false "search content type"
// @Param program query string false "search content program"
// @Param content_name query string false "search content name"
// @Param program_group query string false "search program group"
// @Param source_type query string false "search content source type"
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_authed [get]
func (s *Server) queryAuthContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := queryCondition(c, op)

	hasPermission, err := model.GetContentPermissionModel().CheckQueryContentPermission(ctx, condition, model.QueryModePublished, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	total, results, err := model.GetContentModel().SearchAuthedContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
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
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := queryCondition(c, op)

	//TODO: add check folder permission
	hasPermission, err := model.GetContentPermissionModel().CheckQueryContentPermission(ctx, condition, model.QueryModePublished, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}
	author := c.Query("author")
	total := 0
	var results []*entity.FolderContentData
	if author == constant.Self {
		total, results, err = model.GetContentModel().SearchUserPrivateFolderContent(ctx, dbo.MustGetDB(ctx), condition, op)
	} else {
		total, results, err = model.GetContentModel().SearchUserFolderContent(ctx, dbo.MustGetDB(ctx), condition, op)
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.FolderContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	case model.ErrInvalidVisibleScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
// @Param source_type query string false "search content source type"
// @Param scope query string false "search content scope"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} BadRequestResponse
// @Router /contents_private [get]
func (s *Server) queryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)

	condition := queryCondition(c, op)

	hasPermission, err := model.GetContentPermissionModel().CheckQueryContentPermission(ctx, condition, model.QueryModePrivate, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	total, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	case model.ErrInvalidVisibleScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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
// @Param content_name query string false "search content name"
// @Param program_group query string false "search program group"
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected, archive)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_pending [get]
func (s *Server) queryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op := s.getOperator(c)

	condition := queryCondition(c, op)

	hasPermission, err := model.GetContentPermissionModel().CheckQueryContentPermission(ctx, condition, model.QueryModePending, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, L(GeneralUnknown))
		return
	}

	total, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
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

func queryCondition(c *gin.Context, op *entity.Operator) da.ContentCondition {

	contentTypeStr := c.Query("content_type")
	//keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	scope := c.Query("scope")
	publish := c.Query("publish_status")
	programs := c.Query("programs")
	path := c.Query("path")
	programGroup := c.Query("program_group")
	condition := da.ContentCondition{
		Author:      parseAuthor(c, op),
		Org:         parseOrg(c, op),
		DirPath:     path,
		OrderBy:     da.NewContentOrderBy(c.Query("order_by")),
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
