package model

import (
	"context"
	"sort"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IReportTeachingLoadModel interface {
	ListTeachingLoadReport(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.ReportListTeachingLoadArgs) (*entity.ReportListTeachingLoadResult, error)
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
	// clean args
	var err error
	if args, err = m.cleanListTeachingLoadReportArgs(ctx, tx, operator, args); err != nil {
		log.Error(ctx, "ListTeachingLoadReport: checker.CheckTeachers: search failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}

	if len(args.TeacherIDs) == 0 {
		log.Error(ctx, "ListTeachingLoadReport: not found teachers",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, nil
	}

	// permission check
	checker := NewReportPermissionChecker(operator)
	ok, err := checker.CheckTeachers(ctx, args.TeacherIDs)
	if err != nil {
		log.Error(ctx, "ListTeachingLoadReport: checker.CheckTeachers: search failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
			log.Any("checker", checker),
		)
		return nil, err
	}
	if !ok {
		log.Error(ctx, "ListTeachingLoadReport: checker.CheckTeachers: check failed",
			log.Strings("teacher_ids", args.TeacherIDs),
			log.Any("args", args),
			log.Any("operator", operator),
			log.Any("checker", checker),
		)
		return nil, constant.ErrForbidden
	}

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

func (m *reportTeachingLoadModel) cleanListTeachingLoadReportArgs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.ReportListTeachingLoadArgs) (*entity.ReportListTeachingLoadArgs, error) {
	if args.SchoolID == "" {
		log.Error(ctx, "ListTeachingLoadReport: require school id",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, constant.ErrInvalidArgs
	}
	if len(args.TeacherIDs) == 0 {
		log.Error(ctx, "ListTeachingLoadReport: require teacher ids",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, constant.ErrInvalidArgs
	}
	if len(args.TeacherIDs) > 1 {
		args.ClassIDs = nil
	}

	switch args.SchoolID {
	case string(entity.ListTeachingLoadReportOptionAll):
		args.SchoolID = ""
		if args.TeacherIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.TeacherIDs = nil
			teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "ListTeachingLoadReport: external.GetTeacherServiceProvider().GetByOrganization: require teacher ids",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("args", args),
				)
				return nil, err
			}
			for _, t := range teachers {
				args.TeacherIDs = append(args.TeacherIDs, t.ID)
			}
		} else if len(args.ClassIDs) > 0 && args.ClassIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.ClassIDs = nil
			classes, err := external.GetClassServiceProvider().GetByUserID(ctx, operator, args.TeacherIDs[0])
			if err != nil {
				return nil, err
			}
			for _, c := range classes {
				args.ClassIDs = append(args.ClassIDs, c.ID)
			}
			m, err := external.GetClassServiceProvider().GetByOrganizationIDs(ctx, operator, []string{operator.OrgID})
			if err != nil {
				return nil, err
			}
			orgClassIDs := make([]string, 0, len(classes))
			for _, cc := range m {
				for _, c := range cc {
					orgClassIDs = append(orgClassIDs, c.ID)
				}
			}
			args.ClassIDs = utils.IntersectAndDeduplicateStrSlice(args.ClassIDs, orgClassIDs)
		}
	case string(entity.ListTeachingLoadReportOptionNoAssigned):
		args.SchoolID = ""
		if args.TeacherIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.TeacherIDs = nil
			teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, operator, operator.OrgID)
			if err != nil {
				log.Error(ctx, "ListTeachingLoadReport: external.GetTeacherServiceProvider().GetByOrganization: get failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("args", args),
				)
				return nil, err
			}
			teacherIDs := make([]string, 0, len(teachers))
			for _, t := range teachers {
				teacherIDs = append(teacherIDs, t.ID)
			}
			teacherIDs = utils.SliceDeduplicationExcludeEmpty(teacherIDs)
			userSchoolsMap, err := external.GetSchoolServiceProvider().GetByUsers(ctx, operator, operator.OrgID, teacherIDs)
			if err != nil {
				log.Error(ctx, "ListTeachingLoadReport: external.GetSchoolServiceProvider().BatchGet: batch get failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Any("args", args),
				)
				return nil, err
			}
			for tid, schools := range userSchoolsMap {
				if len(schools) == 0 {
					args.TeacherIDs = append(args.TeacherIDs, tid)
				}
			}
		} else if len(args.ClassIDs) > 0 && args.ClassIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.ClassIDs = nil
			classes, err := external.GetClassServiceProvider().GetByUserID(ctx, operator, args.TeacherIDs[0])
			if err != nil {
				return nil, err
			}
			for _, c := range classes {
				args.ClassIDs = append(args.ClassIDs, c.ID)
			}
			orgClasses, err := external.GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, operator, operator.OrgID)
			if err != nil {
				return nil, err
			}
			orgClassIDs := make([]string, 0, len(orgClasses))
			for _, c := range orgClasses {
				orgClassIDs = append(orgClassIDs, c.ID)
			}
			args.ClassIDs = utils.IntersectAndDeduplicateStrSlice(args.ClassIDs, orgClassIDs)
		}
	default:
		if args.TeacherIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.TeacherIDs = nil
			teachers, err := external.GetTeacherServiceProvider().GetBySchool(ctx, operator, args.SchoolID)
			if err != nil {
				log.Error(ctx, "ListTeachingLoadReport: external.GetTeacherServiceProvider().GetBySchool: get failed",
					log.Err(err),
					log.String("school_id", args.SchoolID),
					log.Any("operator", operator),
					log.Any("args", args),
				)
				return nil, err
			}
			args.TeacherIDs = make([]string, 0, len(teachers))
			for _, t := range teachers {
				args.TeacherIDs = append(args.TeacherIDs, t.ID)
			}
		} else if len(args.ClassIDs) > 0 && args.ClassIDs[0] == string(entity.ListTeachingLoadReportOptionAll) {
			args.ClassIDs = nil
			classes, err := external.GetClassServiceProvider().GetByUserID(ctx, operator, args.TeacherIDs[0])
			if err != nil {
				return nil, err
			}
			for _, c := range classes {
				args.ClassIDs = append(args.ClassIDs, c.ID)
			}
			m, err := external.GetClassServiceProvider().GetBySchoolIDs(ctx, operator, []string{args.SchoolID})
			if err != nil {
				return nil, err
			}
			schoolClassIDs := make([]string, 0, len(classes))
			for _, cc := range m {
				for _, c := range cc {
					schoolClassIDs = append(schoolClassIDs, c.ID)
				}
			}
			args.ClassIDs = utils.IntersectAndDeduplicateStrSlice(args.ClassIDs, schoolClassIDs)
		}
	}

	args.TeacherIDs = utils.SliceDeduplicationExcludeEmpty(args.TeacherIDs)
	args.ClassIDs = utils.SliceDeduplicationExcludeEmpty(args.ClassIDs)

	return args, nil
}
