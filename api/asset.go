package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) createAsset(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err = s.checkAssets(c.Request.Context(), &data)
	if err != nil {
		log.Error(ctx, "Invalid content type", log.Err(err), log.Any("data", data))
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	cid, err := model.GetContentModel().CreateContent(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, L(Unknown))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest,L(Unknown))
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
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}
func (s *Server) updateAsset(c *gin.Context){
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, L(Unknown))
		return
	}

	err = s.checkAssets(c.Request.Context(), &data)
	if err != nil {
		log.Error(ctx, "Invalid content type", log.Err(err), log.Any("data", data))
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
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) deleteAsset(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := model.GetContentModel().DeleteContent(ctx, tx, cid, op)
		if err != nil{
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
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) getAssetByID(c *gin.Context) {
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
func (s *Server) searchAssets(c *gin.Context){
	ctx := c.Request.Context()
	op := GetOperator(c)
	condition := queryCondition(c, op)

	if condition.ContentType == nil {
		condition.ContentType = []int{entity.ContentTypeAssets}
	}

	key, results, err := model.GetContentModel().SearchContent(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"total": key,
			"list":  results,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) getOperator(c *gin.Context) entity.Operator{
	return entity.Operator{}
}


func responseMsg(msg string) interface{}{
	return gin.H{
		"msg": msg,
	}
}
func buildAssetSearchCondition(c *gin.Context) *entity.SearchAssetCondition{
	PageSize, _ := strconv.Atoi("page_size")
	Page, _ := strconv.Atoi("page")
	rawSearchWord := c.Query("search_words")
	isSelfStr := c.Query("is_self")
	fuzzyQuery := c.Query("fuzzy_query")
	orderBy := c.Query("order_by")

	searchWords := strings.Split(rawSearchWord, " ")
	isSelf, _ := strconv.ParseBool(isSelfStr)

	data := &entity.SearchAssetCondition{
		SearchWords:  searchWords,
		FuzzyQuery: fuzzyQuery,
		IsSelf:  isSelf,
		OrderBy: orderBy,
		PageSize: PageSize,
		Page:     Page,
	}

	return data
}

func (s *Server) checkAssets(ctx context.Context, data *entity.CreateContentRequest) error{
	if !data.ContentType.IsAsset() {
		log.Error(ctx, "Invalid content type", log.Err(entity.ErrInvalidContentType), log.Any("data", data))
		return entity.ErrInvalidContentType
	}
	data.Outcomes = nil
	data.SuggestTime = 0
	return nil
}