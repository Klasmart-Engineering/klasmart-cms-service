package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s *Server) createAsset(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = s.checkAssets(c.Request.Context(), data)
	if err != nil {
		log.Error(ctx, "Invalid content type", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	s.fillAssetsRequest(c.Request.Context(), &data)

	cid, err := model.GetContentModel().CreateContent(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequireContentName:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrRequirePublishScope:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case entity.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidSourceExt:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrTeacherManualExtension:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
func (s *Server) updateAsset(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	cid := c.Param("content_id")
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = s.checkAssets(c.Request.Context(), data)
	if err != nil {
		log.Error(ctx, "Invalid content type", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}
	s.fillAssetsRequest(c.Request.Context(), &data)

	err = model.GetContentModel().UpdateContent(ctx, dbo.MustGetDB(ctx), cid, data, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case model.ErrInvalidResourceID:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
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
	case model.ErrInvalidSourceExt:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case model.ErrTeacherManualExtension:
		c.JSON(http.StatusBadRequest, L(LibraryErrorUnsupported))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func (s *Server) deleteAsset(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
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
		c.JSON(http.StatusConflict, L(GeneralUnknown))
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, L(GeneralUnknown))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func (s *Server) getAssetByID(c *gin.Context) {
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

	result, err := model.GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, result)
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}
func (s *Server) searchAssets(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	condition := s.queryContentCondition(c, op)

	if condition.ContentType == nil {
		condition.ContentType = []int{entity.ContentTypeAssets}
	}

	key, results, err := model.GetContentModel().SearchContent(ctx, dbo.MustGetDB(ctx), &condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"total": key,
			"list":  results,
		})
	default:
		c.JSON(http.StatusInternalServerError, L(GeneralUnknown))
	}
}

func responseMsg(msg string) interface{} {
	return gin.H{
		"msg": msg,
	}
}

func (s *Server) checkAssets(ctx context.Context, data entity.CreateContentRequest) error {
	if !data.ContentType.IsAsset() {
		log.Error(ctx, "Invalid content type", log.Err(entity.ErrInvalidContentType), log.Any("data", data))
		return entity.ErrInvalidContentType
	}
	return nil
}

func (s *Server) fillAssetsRequest(ctx context.Context, data *entity.CreateContentRequest) {
	data.Outcomes = nil
	data.SuggestTime = 0
}
