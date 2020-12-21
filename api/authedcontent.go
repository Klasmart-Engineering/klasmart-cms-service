package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary addAuthedContent
// @ID addAuthedContent
// @Description add authed content to org
// @Accept json
// @Produce json
// @Param content body entity.AddAuthedContentRequest true "add authed content request"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth [post]
func (s *Server) addAuthedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.AddAuthedContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetAuthedContentRecordsModel().Add(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary batchAddAuthedContent
// @ID batchAddAuthedContent
// @Description batch add authed content to org
// @Accept json
// @Produce json
// @Param content body entity.BatchAddAuthedContentRequest true "batch add authed content request"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/batch [post]
func (s *Server) batchAddAuthedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.BatchAddAuthedContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetAuthedContentRecordsModel().BatchAdd(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary deleteAuthedContent
// @ID deleteAuthedContent
// @Description delete authed content to org
// @Accept json
// @Produce json
// @Param content body entity.DeleteAuthedContentRequest true "batch delete authed content request"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth [delete]
func (s *Server) deleteAuthedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.DeleteAuthedContentRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetAuthedContentRecordsModel().Delete(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary batchDeleteAuthedContent
// @ID batchDeleteAuthedContent
// @Description batch delete authed content to org
// @Accept json
// @Produce json
// @Param content body entity.BatchDeleteAuthedContentByOrgsRequest true "batch delete authed content request"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/batch [delete]
func (s *Server) batchDeleteAuthedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	var data entity.BatchDeleteAuthedContentByOrgsRequest
	err := c.ShouldBind(&data)
	if err != nil {
		log.Error(ctx, "create content failed", log.Err(err))
		c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		return
	}

	err = model.GetAuthedContentRecordsModel().BatchDelete(ctx, dbo.MustGetDB(ctx), data, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, "ok")
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary getOrgAuthedContent
// @ID getOrgAuthedContent
// @Description get org authed content list
// @Accept json
// @Produce json
// @Param org_id query string false "org id"
// @Tags content
// @Success 200 {object} entity.AuthedContentRecordInfoResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/org [get]
func (s *Server) getOrgAuthedContent(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	orgID := c.Query("org_id")
	total, records, err := model.GetAuthedContentRecordsModel().SearchDetailsList(ctx, dbo.MustGetDB(ctx), entity.SearchAuthedContentRequest{
		OrgIDs: []string{orgID},
	}, op)
	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.AuthedContentRecordInfoResponse{
			Total: total,
			List:  records,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}

// @Summary getContentAuthedOrg
// @ID getContentAuthedOrg
// @Description get content authed org list
// @Accept json
// @Produce json
// @Param content_id query string false "content id"
// @Tags content
// @Success 200 {object} entity.AuthedOrgList
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/content [get]
func (s *Server) getContentAuthedOrg(c *gin.Context) {
	ctx := c.Request.Context()
	op := s.getOperator(c)
	contentID := c.Query("content_id")
	total, records, err := model.GetAuthedContentRecordsModel().SearchRecordsList(ctx, dbo.MustGetDB(ctx), entity.SearchAuthedContentRequest{
		ContentIDs: []string{contentID},
	}, op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	oids := make([]string, len(records))
	for i := range records {
		oids[i] = records[i].OrgID
	}
	orgs, err := external.GetOrganizationServiceProvider().BatchGet(ctx, op, oids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	list := make([]*entity.OrganizationInfo, len(orgs))
	for i := range orgs {
		list[i] = &entity.OrganizationInfo{
			ID:   orgs[i].ID,
			Name: orgs[i].Name,
		}
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, entity.AuthedOrgList{
			Total: total,
			List:  list,
		})
	default:
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
	}
}
