package api

import (
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

func (s *Server) createContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	cid, err := model.GetContentModel().CreateContent(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": cid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) publishContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	err = model.GetContentModel().PublishContentBulk(ctx, dbo.MustGetDB(ctx), ids.ID, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) publishContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data struct {
		Scope string `json:"scope"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	err = model.GetContentModel().PublishContent(ctx, dbo.MustGetDB(ctx), cid, data.Scope, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) GetContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data struct {
		Scope string `json:"scope"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
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

func (s *Server) updateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	err = model.GetContentModel().UpdateContent(ctx, dbo.MustGetDB(ctx), cid, data, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrNoAuth:
		c.JSON(http.StatusForbidden, responseMsg(err.Error()))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) lockContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	ncid, err := model.GetContentModel().LockContent(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
	case model.ErrInvalidPublishStatus:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrInvalidContentType:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
	case model.ErrContentAlreadyLocked:
		c.JSON(http.StatusNotAcceptable, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"content_id": ncid,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) deleteContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	ids := new(contentBulkOperateRequest)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	err = model.GetContentModel().DeleteContentBulk(ctx, dbo.MustGetDB(ctx), ids.ID, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) deleteContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")

	err := model.GetContentModel().DeleteContent(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

func (s *Server) QueryDynamoContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	ctoips := ""
	if c.Query("content_type") != "" {
		ctoips = c.Query("content_type") + c.Query("org") + c.Query("publish_status")
	}

	condition := da.DyKeyContentCondition{
		PublishStatus:                 c.Query("publish_status"),
		Author:                        c.Query("author"),
		ContentTypeOrgIdPublishStatus: ctoips,
		Name:                          c.Query("name"),
		Org:                           c.Query("org"),
		KeyWords:                      c.Query("keywords"),
		LastKey:                       c.Query("key"),
	}

	key, results, err := model.GetContentModel().SearchContentByDynamoKey(ctx, dbo.MustGetDB(ctx), condition, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"key":  key,
			"list": results,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}

}

func (s *Server) QueryContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	condition := queryCondition(c, op)
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

func (s *Server) QueryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	key, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
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

func (s *Server) QueryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	key, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), condition, op)
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

	contentType, _ := strconv.Atoi(c.Query("content_type"))
	//keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	scope := c.Query("scope")
	publish := c.Query("publish_status")
	condition := da.ContentCondition{
		Author:  parseAuthor(c, op),
		Org:     parseOrg(c, op),
		OrderBy: da.NewContentOrderBy(c.Query("order_by")),
		Pager:   utils.GetPager(c.Query("page"), c.Query("page_size")),
		Name:    strings.TrimSpace(c.Query("name")),
	}
	//if len(keywords) > 0 {
	//	condition.Name = keywords
	//}
	if contentType != 0 {
		condition.ContentType = append(condition.ContentType, contentType)
	}
	if scope != "" {
		condition.Scope = append(condition.Scope, scope)
	}
	if publish != "" {
		condition.PublishStatus = append(condition.PublishStatus, publish)
	}
	return condition
}
