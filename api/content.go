package api

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s *Server) createContent(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
	var data entity.CreateContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	cid, err := model.GetContentModel().CreateContent(ctx, dbo.MustGetDB(ctx), data, op)

	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"content_id": cid,
	})
}

func (s *Server) publishContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
	cid := c.Param("content_id")
	var data struct {
		Scope string `json:"scope"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, responseMsg(err.Error()))
		return
	}

	result, err := model.GetContentModel().GetContentByID(ctx, dbo.MustGetDB(ctx), cid, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) updateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *Server) lockContent(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
		"content_id": ncid,
	})
}

func (s *Server) deleteContentBulk(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
	cid := c.Param("content_id")

	err := model.GetContentModel().DeleteContent(ctx, dbo.MustGetDB(ctx), cid, op)
	switch err{
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
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}
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
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}

	contentType, _ := strconv.Atoi(c.Query("content_type"))

	keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	condition := da.ContentCondition{
		Name:          keywords,
		ContentType:   []int{contentType},
		PublishStatus: []string{c.Query("publish_status")},
		Scope:         []string{c.Query("scope")},
		Author:        c.Query("author"),
		Org:           c.Query("org"),
		OrderBy:       da.NewContentOrderBy(c.Query("order_by")),
		Pager:			utils.GetPager(c.Query("page"),c.Query("page_size")),
	}

	key, results, err := model.GetContentModel().SearchContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"key":  key,
		"list": results,
	})
}

func (s *Server) QueryPrivateContent(c *gin.Context) {
	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}

	contentType, _ := strconv.Atoi(c.Query("content_type"))
	keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	condition := da.ContentCondition{
		Name:         keywords,
		ContentType:   []int{contentType},
		PublishStatus: []string{c.Query("publish_status")},
		Scope:         []string{c.Query("scope")},
		Author:        c.Query("author"),
		Org:           c.Query("org"),
		OrderBy:      da.NewContentOrderBy(c.Query("order_by")),
		Pager:			utils.GetPager(c.Query("page"),c.Query("page_size")),
	}

	key, results, err := model.GetContentModel().SearchUserPrivateContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"key":  key,
		"list": results,
	})
}

func (s *Server) QueryPendingContent(c *gin.Context) {

	ctx := c.Request.Context()
	op, exist := GetOperator(c)
	if !exist {
		c.JSON(http.StatusUnauthorized, "get operator failed")
		return
	}

	contentType, _ := strconv.Atoi(c.Query("content_type"))
	keywords := strings.Split(strings.TrimSpace(c.Query("name")), " ")
	condition := da.ContentCondition{
		Name:          keywords,
		ContentType:   []int{contentType},
		PublishStatus: []string{c.Query("publish_status")},
		Scope:         []string{c.Query("scope")},
		Author:        c.Query("author"),
		Org:           c.Query("org"),
		OrderBy:      da.NewContentOrderBy(c.Query("order_by")),
		Pager:			utils.GetPager(c.Query("page"),c.Query("page_size")),
	}

	key, results, err := model.GetContentModel().ListPendingContent(ctx, dbo.MustGetDB(ctx), condition, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"key":  key,
		"list": results,
	})
}
