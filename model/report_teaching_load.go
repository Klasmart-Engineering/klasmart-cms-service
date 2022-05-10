package model

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/da"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type IReportTeachingLoadModel interface {
	ListTeachingLoadReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.ReportListTeachingLoadArgs) (*entity.ReportListTeachingLoadResult, error)
	GetTeacherLoadOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, teacherIDs []string, classIDs []string) (res *entity.TeacherLoadOverview, err error)
}

var (
	reportTeachingLoadModelInstance IReportTeachingLoadModel
	reportTeachingLoadModelOnce     = sync.Once{}
)

func GetReportTeachingLoadModel() IReportTeachingLoadModel {
	reportTeachingLoadModelOnce.Do(func() {
		reportTeachingLoadModelInstance = &reportTeachingLoadModel{}
	})
	return reportTeachingLoadModelInstance
}

type reportTeachingLoadModel struct{}

func (m *reportTeachingLoadModel) ListTeachingLoadReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.ReportListTeachingLoadArgs) (*entity.ReportListTeachingLoadResult, error) {
	var err error

	// prepend time ranges
	var (
		ranges     = make([]*entity.ScheduleTimeRange, 0, constant.ReportTeachingLoadDays)
		loc        = time.FixedZone("report_teaching_load", args.TimeOffset)
		now        = time.Now()
		start, end = utils.BeginOfDayByTime(now, loc), utils.EndOfDayByTime(now, loc)
	)
	for i := 0; i < constant.ReportTeachingLoadDays; i++ {
		ranges = append(ranges, &entity.ScheduleTimeRange{
			StartAt: start.Unix(),
			EndAt:   end.Unix(),
		})
		start = start.AddDate(0, 0, 1)
		end = end.AddDate(0, 0, 1)
	}

	// call schedule module
	input := entity.ScheduleTeachingLoadInput{
		OrgID:      operator.OrgID,
		ClassIDs:   args.ClassIDs,
		TeacherIDs: args.TeacherIDs,
		TimeRanges: ranges,
	}
	log.Debug(ctx, "ListTeachingLoadReport: print call schedule args",
		log.Any("input", input),
	)
	loads, err := GetScheduleModel().GetTeachingLoad(ctx, &input)
	if err != nil {
		log.Error(ctx, "ListTeachingLoadReport: GetScheduleModel().GetTeachingLoad: get failed",
			log.Err(err),
			log.Any("input", input),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// batch get teacher name map
	teacherNameMap, err := external.GetTeacherServiceProvider().BatchGetNameMap(ctx, operator, args.TeacherIDs)
	if err != nil {
		log.Error(ctx, "ListTeachingLoadReport: external.GetTeacherServiceProvider().BatchGetNameMap: batch get failed",
			log.Err(err),
			log.Any("teacher_ids", args.TeacherIDs),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// build duration map and teacher load map
	type durationKey struct {
		TeacherID string
		ClassType entity.ScheduleClassType
		StartAt   int64
		EndAt     int64
	}
	durationsMap := make(map[durationKey]int64, len(loads)*constant.ReportTeachingLoadDays)
	teacherLoadMap := make(map[string]*entity.ScheduleTeachingLoadView, len(args.TeacherIDs))
	for _, l := range loads {
		teacherLoadMap[l.TeacherID] = l
		for _, d := range l.Durations {
			durationsMap[durationKey{
				TeacherID: l.TeacherID,
				ClassType: l.ClassType,
				StartAt:   d.StartAt,
				EndAt:     d.EndAt,
			}] = d.Duration
		}
	}

	// mapping result
	r := entity.ReportListTeachingLoadResult{
		Items: make([]*entity.ReportListTeachingLoadItem, 0, len(args.TeacherIDs)),
		Total: len(args.TeacherIDs),
	}
	for _, tid := range args.TeacherIDs {
		l := teacherLoadMap[tid]
		if l == nil {
			l = &entity.ScheduleTeachingLoadView{TeacherID: tid}
		}
		item := entity.ReportListTeachingLoadItem{
			TeacherID:   l.TeacherID,
			TeacherName: teacherNameMap[l.TeacherID],
		}
		for _, r := range ranges {
			online := durationsMap[durationKey{
				TeacherID: l.TeacherID,
				ClassType: entity.ScheduleClassTypeOnlineClass,
				StartAt:   r.StartAt,
				EndAt:     r.EndAt,
			}]
			offline := durationsMap[durationKey{
				TeacherID: l.TeacherID,
				ClassType: entity.ScheduleClassTypeOfflineClass,
				StartAt:   r.StartAt,
				EndAt:     r.EndAt,
			}]
			item.Durations = append(item.Durations, &entity.ReportListTeachingLoadDuration{
				StartAt: r.StartAt,
				EndAt:   r.EndAt,
				Online:  online,
				Offline: offline,
			})
		}
		r.Items = append(r.Items, &item)
	}

	sort.Sort(entity.ReportListTeachingLoadItemsSortByTeacherName(r.Items))

	return &r, nil
}

func (m *reportTeachingLoadModel) GetTeacherLoadOverview(ctx context.Context, op *entity.Operator, tr entity.TimeRange, teacherIDs []string, classIDs []string) (res *entity.TeacherLoadOverview, err error) {
	res = &entity.TeacherLoadOverview{}
	items, err := da.GetReportDA().GetTeacherLoadItems(ctx, op, tr, teacherIDs, classIDs)
	if err != nil {
		return
	}

	for _, item := range items {
		if item.TotalLessons == 0 {
			continue
		}
		attendedPercent := (float64(item.TotalLessons) - float64(item.MissedLessons)) / float64(item.TotalLessons)
		switch {
		case attendedPercent == 1:
			res.NumOfTeachersCompletedAll++
		case attendedPercent >= 0.8 && attendedPercent < 1:
			res.NumOfTeachersMissedSome++
		default:
			res.NumOfTeachersMissedFrequently++
		}
		res.NumOfMissedLessons += item.MissedLessons
	}
	return
}
