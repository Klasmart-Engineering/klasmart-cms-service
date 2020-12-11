package api

import "github.com/gin-gonic/gin"

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

}

// @Summary deleteAuthedContent
// @ID deleteAuthedContent
// @Description batch delete authed content to org
// @Accept json
// @Produce json
// @Param content body entity.BatchDeleteAuthedContentRequest true "batch delete authed content request"
// @Tags content
// @Success 200 {object} string
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth [delete]
func (s *Server) batchDeleteAuthedContent(c *gin.Context) {

}

// @Summary getOrgAuthedContent
// @ID getOrgAuthedContent
// @Description get org authed content list
// @Accept json
// @Produce json
// @Param org_id query string false "org id"
// @Tags content
// @Success 200 {object} AuthedContentRecordInfoResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/org [get]
func (s *Server) getOrgAuthedContent(c *gin.Context) {

}

// @Summary getContentAuthedOrg
// @ID getContentAuthedOrg
// @Description get content authed org list
// @Accept json
// @Produce json
// @Param content_id query string false "content id"
// @Tags content
// @Success 200 {object} AuthedOrgList
// @Failure 500 {object} InternalServerErrorResponse
// @Failure 400 {object} BadRequestResponse
// @Router /contents_auth/content [get]
func (s *Server) getContentAuthedOrg(c *gin.Context) {

}
