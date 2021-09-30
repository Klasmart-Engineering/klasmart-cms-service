package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (t *reportModel) ListTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) ([]*entity.TeacherLoadLesson, error) {
	mapTeacherClassWithStudent, err := external.GetTeacherLoadServiceProvider().BatchGetClassWithStudent(ctx, op, args.TeacherIDs)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: call ams failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("args", args))
		return nil, err
	}
	res, err := da.GetReportDA().ListTeacherLoadLessons(ctx, op, dbo.MustGetDB(ctx), args)
	if err != nil {
		log.Error(ctx, "ListTeacherLoadLessons: call da failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("args", args))
		return nil, err
	}
	mapTeacherLoadLesson := make(map[string]*entity.TeacherLoadLesson, len(res))
	for i := range res {
		mapTeacherLoadLesson[res[i].TeacherID] = res[i]
	}

	result := make([]*entity.TeacherLoadLesson, len(args.TeacherIDs))
	for i, tid := range args.TeacherIDs {
		load := entity.TeacherLoadLesson{TeacherID: tid}
		if mapTeacherLoadLesson[tid] != nil {
			load.CompletedLiveLessons = mapTeacherLoadLesson[tid].CompletedLiveLessons
			load.CompletedInClassLessons = mapTeacherLoadLesson[tid].CompletedInClassLessons
			load.MissedLiveLessons = mapTeacherLoadLesson[tid].MissedLiveLessons
			load.MissedInClassLessons = mapTeacherLoadLesson[tid].MissedInClassLessons
			load.TotalScheduled = mapTeacherLoadLesson[tid].TotalScheduled
		}
		if mapTeacherClassWithStudent[tid] != nil {
			counter := mapTeacherClassWithStudent[tid].CountClassAndStudent(ctx, args.ClassIDs)
			load.NumberOfClasses = counter.Class
			load.NumberOfStudents = counter.Student
		}
		result[i] = &load
	}
	return result, nil
}

func (t *reportModel) SummaryTeacherLoadLessons(ctx context.Context, op *entity.Operator, args *entity.TeacherLoadLessonArgs) (*entity.TeacherLoadLessonSummary, error) {
	result, err := da.GetReportDA().SummaryTeacherLoadLessons(ctx, op, dbo.MustGetDB(ctx), args)
	if err != nil {
		log.Error(ctx, "SummaryTeacherLoadLessons: call da failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("args", args))
		return nil, err
	}
	return &entity.TeacherLoadLessonSummary{
		CompletedLiveLessons:    entity.SummaryNode{Count: result.LiveCompletedCount, Duration: result.LiveCompletedDuration},
		CompletedInClassLessons: entity.SummaryNode{Count: result.InClassCompletedCount, Duration: result.InClassCompletedDuration},
		MissedLiveLessons:       entity.SummaryNode{Count: result.LiveMissedCount, Duration: result.LiveMissedDuration},
		MissedInClassLessons:    entity.SummaryNode{Count: result.InClassMissedCount, Duration: result.InClassMissedDuration},
	}, nil
}
func (t *reportModel) MissedLessonsList(ctx context.Context, request *entity.TeacherLoadMissedLessonsRequest) (response *entity.TeacherLoadMissedLessonsResponse, err error) {
	response = new(entity.TeacherLoadMissedLessonsResponse)
	reportDA := da.GetReportDA()
	list, err := reportDA.MissedLessonsListInfo(ctx, request)
	if err != nil {
		return nil, err
	}
	total, err := reportDA.MissedLessonsListTotal(ctx, request)
	if err != nil {
		return nil, err
	}
	response.List = list
	response.Total = total
	return
}
