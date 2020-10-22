package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
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
type PublishContentRequest struct {
	Scope string `json:"scope"`
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
	op := GetOperator(c)
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	cid, err := model.GetContentModel().CreateContent(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidSelectForm:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := model.GetContentModel().PublishContentBulk(ctx, tx, ids.ID, op)
		if err != nil {
			return err
		}
		return nil
	})
	switch err {
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	cid := c.Param("content_id")

	data := new(PublishContentRequest)
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err = model.GetContentModel().PublishContent(ctx, dbo.MustGetDB(ctx), cid, data.Scope, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	cid := c.Param("content_id")

	data := new(PublishContentRequest)
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err = model.GetContentModel().PublishContentWithAssets(ctx, dbo.MustGetDB(ctx), cid, data.Scope, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}


// @Summary getContent
// @ID getContentById
// @Description get a content by id
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
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data struct {
		Scope string `json:"scope"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	result, err := model.GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = model.GetContentModel().UpdateContent(ctx, dbo.MustGetDB(ctx), cid, data, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, L(Unknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	cid := c.Param("content_id")
	ncid, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		ncid, err := model.GetContentModel().LockContent(ctx, tx, cid, op)
		if err != nil {
			return nil, err
		}
		return ncid, nil
	})
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(Unknown))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrContentAlreadyLocked:
		c.JSON(http.StatusNotAcceptable, L(Unknown))
	case model.ErrInvalidLockedContentPublishStatus:
		c.JSON(http.StatusConflict, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": ncid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)

	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err = model.GetContentModel().DeleteContentBulk(ctx, tx, ids.ID, op)
		if err != nil {
			return err
		}
		return nil
	})
	switch err {
	case model.ErrDeleteLessonInSchedule:
		c.JSON(http.StatusConflict, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
	op := GetOperator(c)
	cid := c.Param("content_id")

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := model.GetContentModel().DeleteContent(ctx, tx, cid, op)
		if err != nil {
			return err
		}
		return nil
	})
	switch err {
	case model.ErrDeleteLessonInSchedule:
		c.JSON(http.StatusConflict, L(Unknown))
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(Unknown))
	case nil:
		c.JSON(http.StatusOK, "")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary queryContent
// @ID searchContents
// @Description query content by condition
// @Accept json
// @Produce json
// @Param name query string false "search content name"
// @Param author query string false "search content author"
// @Param content_type query string false "search content type"
// @Param scope query string false "search content scope"
// @Param program query string false "search content program"
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
	op := GetOperator(c)
	condition := queryCondition(c, op)
	total, results, err := model.GetContentModel().SearchContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Param source_type query string false "search content source type"
// @Param scope query string false "search content scope"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/private [get]
func (s *Server) queryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	total, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
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
// @Param source_type query string false "search content source type"
// @Param publish_status query string  false "search content publish status" Enums(published, draft, pending, rejected)
// @Param order_by query string false "search content order by column name" Enums(id, -id, content_name, -content_name, create_at, -create_at, update_at, -update_at)
// @Param page_size query int false "content list page size"
// @Param page query int false "content list page index"
// @Tags content
// @Success 200 {object} entity.ContentInfoWithDetailsResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents/pending [get]
func (s *Server) queryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	total, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &entity.ContentInfoWithDetailsResponse{
			Total:       total,
			ContentList: results,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func parseAuthor(c *gin.Context, u *entity.Operator) string {
	author := c.Query("author")
	if author == "{self}" {
		author = u.UserID
	}
	return author
}

func parseOrg(c *gin.Context, u *entity.Operator) string {
	author := c.Query("org")
	if author == "{self}" {
		author = u.OrgID
	}
	return author
}

func queryCondition(c *gin.Context, op *entity.Operator) da.ContentCondition {

	contentTypeStr := c.Query("content_type")
	//keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	scope := c.Query("scope")
	publish := c.Query("publish_status")
	program := c.Query("program")
	condition := da.ContentCondition{
		Author:  parseAuthor(c, op),
		Org:     parseOrg(c, op),
		OrderBy: da.NewContentOrderBy(c.Query("order_by")),
		Pager:   utils.GetPager(c.Query("page"), c.Query("page_size")),
		Name:    strings.TrimSpace(c.Query("name")),
	}
	sourceType := c.Query("source_type")
	//if len(keywords) > 0 {
	//	condition.Name = keywords
	//}
	if contentTypeStr != "" {
		contentTypeList := strings.Split(contentTypeStr, ",")
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
		scopes := strings.Split(scope, ",")
		condition.Scope = append(condition.Scope, scopes...)
	}
	if publish != "" {
		condition.PublishStatus = append(condition.PublishStatus, publish)
	}
	if program != "" {
		condition.Program = program
	}
	if sourceType != "" {
		condition.SourceType = sourceType
	}
	return condition
}
