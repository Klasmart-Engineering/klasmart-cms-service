package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IReportModel interface {
	ListStudentReport(ctx context.Context, tx *dbo.DBContext, lessonPlanID string, operator entity.Operator) (entity.StudentReportList, error)
	GetStudentReportDetail(ctx context.Context, tx *dbo.DBContext, studentID string, lessonPlanID string, operator entity.Operator) (entity.StudentReportDetail, error)
	GetReportLessPlanInfo(ctx context.Context, tx *dbo.DBContext, teacherID string, classID string, operator entity.Operator) ([]*entity.ReportLessonPlanInfo, error)
}

var (
	reportModelInstance IReportModel
	reportModelOnce     = sync.Once{}
)

func GetReportModel() IReportModel {
	reportModelOnce.Do(func() {
		reportModelInstance = &reportModel{}
	})
	return reportModelInstance
}

type reportModel struct{}

func (r *reportModel) GetReportLessPlanInfo(ctx context.Context, tx *dbo.DBContext, teacherID string, classID string, operator entity.Operator) ([]*entity.ReportLessonPlanInfo, error) {
	lessonPlanIDs, err := da.GetScheduleDA().GetLessonPlanIDsByTeacherAndClass(ctx, tx, teacherID, classID)
	if err != nil {
		logger.Error(ctx, "GetLessPlanInfo:get lessonPlanIDs error",
			log.Err(err),
			log.String("teacherID", teacherID),
			log.String("classID", classID),
			log.Any("operator", operator),
		)
		return nil, err
	}
	lessonPlanInfos, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		logger.Error(ctx, "GetLessPlanInfo:get lessonPlan info error",
			log.Err(err),
			log.String("teacherID", teacherID),
			log.String("classID", classID),
			log.Strings("lessonPlanIDs", lessonPlanIDs),
			log.Any("operator", operator),
		)
	}
	result := make([]*entity.ReportLessonPlanInfo, len(lessonPlanInfos))
	for i, item := range lessonPlanInfos {
		result[i] = &entity.ReportLessonPlanInfo{
			ID:   item.ID,
			Name: item.Name,
		}
	}
	return result, nil
}

func (r *reportModel) ListStudentReport(ctx context.Context, tx *dbo.DBContext, lessonPlanID string, operator entity.Operator) (entity.StudentReportList, error) {
	panic("implement me")
}

func (r *reportModel) GetStudentReportDetail(ctx context.Context, tx *dbo.DBContext, studentID string, lessonPlanID string, operator entity.Operator) (entity.StudentReportDetail, error) {
	panic("implement me")
}
