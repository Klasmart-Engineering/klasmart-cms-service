package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"strconv"
	"strings"
)

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
		return
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": cid,
	})
}

func (s *Server) publishContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	ids := make([]string, 0)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	err = model.GetContentModel().PublishContentBulk(ctx, dbo.MustGetDB(ctx), ids, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
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
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, result)
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
		return
	case model.ErrInvalidResourceId:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	case model.ErrResourceNotFound:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	case model.ErrNoContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	case model.ErrInvalidContentData:
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *Server) lockContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")
	ncid, err := model.GetContentModel().LockContent(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrNoContent:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": ncid,
	})
}

func (s *Server) deleteContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	ids := make([]string, 0)
	err := c.ShouldBind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	err = model.GetContentModel().DeleteContentBulk(ctx, dbo.MustGetDB(ctx), ids, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *Server) deleteContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	cid := c.Param("content_id")

	err := model.GetContentModel().DeleteContent(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err {
	case model.ErrReadContentFailed:
		c.JSON(http.StatusNotFound, responseMsg(err.Error()))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"key":  key,
		"list": results,
	})
}

func (s *Server) QueryContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)
	condition := queryCondition(c, op)
	key, results, err := model.GetContentModel().SearchContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": key,
		"list":  results,
	})
}

func (s *Server) QueryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	key, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": key,
		"list":  results,
	})
}

func (s *Server) QueryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op := GetOperator(c)

	condition := queryCondition(c, op)
	key, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": key,
		"list":  results,
	})
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
		Name:  	 strings.TrimSpace(c.Query("name")),
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
