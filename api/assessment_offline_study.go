package api

// @Summary query user offline study(homeFun study)
// @Description query user offline study
// @Tags assessments
// @ID queryUserOfflineStudy
// @Accept json
// @Produce json
// @Param status query string false "status search"
// @Param query_key query string false "query key fuzzy search"
// @Param query_type query string false "query type" enums(TeacherName)
// @Param page query int false "page number" default(1)
// @Param page_size query integer false "page size" format(int) default(10)
// @Param order_by query string false "query order by" enums(complete_at,-complete_at,submit_at,-submit_at) default(-submit_at)
// @Success 200 {object} v2.OfflineStudyUserPageReply
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /user_offline_study [get]
//func (s *Server) queryUserOfflineStudy(c *gin.Context) {
//	ctx := c.Request.Context()
//	op := s.getOperator(c)
//	req := new(v2.AssessmentQueryReq)
//	if err := c.ShouldBind(req); err != nil {
//		return
//	}
//
//	if req.PageSize <= 0 || req.PageIndex <= 0 {
//		req.PageIndex = constant.DefaultPageIndex
//		req.PageSize = constant.DefaultPageSize
//	}
//
//	log.Debug(ctx, "queryUserOfflineStudy request", log.Any("req", req), log.Any("op", op))
//
//	result, err := model.GetAssessmentOfflineStudyModel().Page(ctx, op, req)
//	switch err {
//	case nil:
//		c.JSON(http.StatusOK, result)
//	case constant.ErrForbidden:
//		c.JSON(http.StatusForbidden, L(AssessMsgNoPermission))
//	default:
//		s.defaultErrorHandler(c, err)
//	}
//}
