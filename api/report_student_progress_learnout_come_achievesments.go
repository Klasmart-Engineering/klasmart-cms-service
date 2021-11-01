package api

import (
	"context"
	"net/http"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

// @Summary  getLearnOutcomeAchievement
// @Tags reports/studentProgress
// @ID getLearnOutcomeAchievement
// @Accept json
// @Produce json
// @Param request body entity.LearnOutcomeAchievementRequest true "request "
// @Success 200 {object} entity.LearnOutcomeAchievementResponse
// @Failure 400 {object} BadRequestResponse
// @Failure 403 {object} ForbiddenResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /reports/student_progress/learn_outcome_achievement [post]
func (s *Server) getLearnOutcomeAchievement(c *gin.Context) {
	ctx := c.Request.Context()
	var err error
	defer func() {
		if err == nil {
			return
		}
		switch err {
		case constant.ErrInvalidArgs:
			c.JSON(http.StatusBadRequest, L(GeneralUnknown))
		case constant.ErrForbidden:
			c.JSON(http.StatusForbidden, L(GeneralUnknown))
		default:
			s.defaultErrorHandler(c, err)
		}
	}()
	op := s.getOperator(c)
	req := entity.LearnOutcomeAchievementRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		log.Error(ctx, "invalid request", log.Err(err))
		err = constant.ErrInvalidArgs
		return
	}
	for _, duration := range req.Durations {
		_, _, err = duration.Value(ctx)
		if err != nil {
			return
		}
	}

	err = s.checkPermissionForReportStudentProgress(ctx, op, req.ClassID, req.StudentID)
	if err != nil {
		return
	}
	res, err := model.GetReportModel().GetStudentProgressLearnOutcomeAchievement(ctx, op, &req)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) checkPermissionForReportStudentProgress(ctx context.Context, op *entity.Operator, classID, studentID string) (err error) {
	permissions, err := external.GetPermissionServiceProvider().HasOrganizationPermissions(ctx, op, []external.PermissionName{
		external.ReportStudentProgressReportView,
		external.ReportStudentProgressReportOrganization,
		external.ReportStudentProgressReportSchool,
		external.ReportStudentProgressReportTeacher,
		external.ReportStudentProgressReportStudent,
	})
	if !permissions[external.ReportStudentProgressReportView] {
		log.Error(ctx, "op has no permission", log.Any("perm", external.ReportStudentProgressReportView))
		err = constant.ErrForbidden
		return
	}
	mOrg, err := external.GetOrganizationServiceProvider().GetByClasses(ctx, op, []string{classID})
	if err != nil {
		return
	}
	org, ok := mOrg[classID]
	if !ok || org.ID != op.OrgID {
		log.Error(ctx, "class not belongs to organization",
			log.Any("classID", classID),
			log.Any("orgID", op.OrgID),
		)
		err = constant.ErrForbidden
		return
	}

	students, err := external.GetStudentServiceProvider().GetByClassID(ctx, op, classID)
	if err != nil {
		return
	}
	isStudentInClass := false
	for _, student := range students {
		if student.ID == studentID {
			isStudentInClass = true
			break
		}
	}
	if !isStudentInClass {
		log.Error(ctx, "student not in class ",
			log.Any("student_id", studentID),
			log.Any("class_id", classID),
			log.Any("students", students),
		)
		err = constant.ErrForbidden
		return
	}

	if permissions[external.ReportStudentProgressReportOrganization] {
		return
	}

	if permissions[external.ReportStudentProgressReportSchool] {
		var mSchools map[string][]*external.School
		mSchools, err = external.GetSchoolServiceProvider().GetByUsers(ctx, op, op.OrgID, []string{op.UserID})
		if err != nil {
			return
		}
		userSchools := mSchools[op.UserID]

		var mClassSchools map[string][]*external.School
		mClassSchools, err = external.GetSchoolServiceProvider().GetByClasses(ctx, op, []string{classID})
		if err != nil {
			return
		}
		classSchools := mClassSchools[classID]

		if len(userSchools) == 0 && len(classSchools) == 0 {
			// op and classID belongs to organization insteadof any school
			return
		}

		for _, userSchool := range userSchools {
			for _, classSchool := range classSchools {
				if userSchool.ID == classSchool.ID {
					return
				}
			}
		}
	}

	if permissions[external.ReportStudentProgressReportTeacher] {
		var classes []*external.Class
		classes, err = external.GetClassServiceProvider().GetByUserID(ctx, op, op.UserID)
		if err != nil {
			return
		}
		for _, class := range classes {
			if class.JoinType == external.IsTeaching && class.ID == classID {
				return
			}
		}
	}

	if permissions[external.ReportStudentProgressReportStudent] {
		if studentID == op.UserID {
			return
		}
	}

	log.Info(ctx, "no permission for checkPermissionForReportStudentProgress",
		log.Any("op", op),
		log.Any("class_id", classID),
		log.Any("student_id", studentID),
		log.Any("permissions", permissions),
	)
	err = constant.ErrForbidden
	return
}
