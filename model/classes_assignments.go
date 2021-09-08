package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IClassesAssignments interface {
	CreateRecord(ctx context.Context, op *entity.Operator, record *entity.AddClassAndLiveAssessmentArgs) error
	GetOverview(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentOverViewRequest) ([]*entity.ClassesAssignmentOverView, error)
	GetStatistic(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentsViewRequest) ([]*entity.ClassesAssignmentsView, error)
	GetUnattended(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentsUnattendedViewRequest) ([]*entity.ClassesAssignmentsUnattendedStudentsView, error)
}

var (
	_classesAssignmentsModel     IClassesAssignments
	_classesAssignmentsModelOnce sync.Once
)

type ClassesAssignmentsModel struct {
}

func (c ClassesAssignmentsModel) CreateRecord(ctx context.Context, op *entity.Operator, data *entity.AddClassAndLiveAssessmentArgs) error {
	schedule, err := GetScheduleModel().GetPlainByID(ctx, data.ScheduleID)
	if err != nil {
		log.Error(ctx, "CreateRecord: GetPlainByID failed", log.Err(err), log.Any("data", data))
		return err
	}
	classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, schedule.ID)
	if err != nil {
		log.Error(ctx, "CreateRecord: GetClassRosterID failed", log.Err(err), log.Any("data", data))
		return err
	}
	if classID == "" {
		log.Info(ctx, "CreateRecord: schedule doesn't belong any class", log.Any("data", data))
		return nil
	}

	attendances := utils.SliceDeduplicationExcludeEmpty(data.AttendanceIDs)
	records := make([]*entity.ClassesAssignmentsRecords, 0, len(attendances))

	for i := range attendances {
		record := entity.ClassesAssignmentsRecords{
			ID:              utils.NewID(),
			ClassID:         classID,
			ScheduleID:      schedule.ID,
			AttendanceID:    attendances[i],
			ScheduleType:    entity.NewScheduleInReportType(schedule.ClassType, schedule.IsHomeFun),
			ScheduleStartAt: schedule.StartAt,
			ScheduleEndAt:   data.ClassEndTime,
			CreateAt:        time.Now().Unix(),
		}
		records = append(records, &record)
	}

	panic("implement da")
}

func (c ClassesAssignmentsModel) GetOverview(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentOverViewRequest) ([]*entity.ClassesAssignmentOverView, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationIDs:  entity.NullStrings{Strings: request.ClassIDs, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterClass), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetOverview: get class's schedules failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	scheduleIDs := make([]string, len(relations))
	for i := range relations {
		scheduleIDs[i] = relations[i].ScheduleID
	}

	min := int64(^uint64(0) >> 1)
	max := int64(0)
	for i := range request.Durations {
		startAt, endAt, err := request.Durations[i].Value(ctx)
		if err != nil {
			log.Error(ctx, "GetOverview: extract time duration failed", log.Err(err), log.Any("request", request))
			return nil, err
		}
		if min > startAt {
			min = startAt
		}
		if max < endAt {
			max = endAt
		}
	}

	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs:       entity.NullStrings{Strings: scheduleIDs, Valid: true},
		StartAtGe: sql.NullInt64{Int64: min, Valid: true},
		StartAtLt: sql.NullInt64{Int64: max, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetOverview: get class's duration schedules failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	overviews := []*entity.ClassesAssignmentOverView{
		{Type: "", Count: 0},
		{Type: "", Count: 0},
		{Type: "", Count: 0},
	}
	for i := range schedules {
		if schedules[i].ClassType == entity.ScheduleClassTypeOnlineClass {
			overviews[0].Count++
		}
		if schedules[i].ClassType == entity.ScheduleClassTypeHomework && !schedules[i].IsHomeFun {
			overviews[1].Count++
		}
		if schedules[i].ClassType == entity.ScheduleClassTypeHomework && schedules[i].IsHomeFun {
			overviews[2].Count++
		}
	}
	return overviews, nil
}

func (c ClassesAssignmentsModel) getTimeRangeSchedule(ctx context.Context, relations []*entity.ScheduleRelation, durations []entity.TimeRange, kind string) (map[entity.TimeRange][]string, error) {
	ids := make([]string, len(relations))
	for i := range relations {
		ids[i] = relations[i].ScheduleID
	}
	min := int64(^uint64(0) >> 1)
	max := int64(0)
	for i := range durations {
		startAt, endAt, err := durations[i].Value(ctx)
		if err != nil {
			log.Error(ctx, "getTimeRangeSchedule: extract time duration failed", log.Err(err))
			return nil, err
		}
		if min > startAt {
			min = startAt
		}
		if max < endAt {
			max = endAt
		}
	}
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs:        entity.NullStrings{Strings: ids, Valid: true},
		StartAtGe:  sql.NullInt64{Int64: min, Valid: true},
		StartAtLt:  sql.NullInt64{Int64: max, Valid: true},
		ClassTypes: entity.NullStrings{Strings: []string{kind}, Valid: true},
		IsHomefun:  sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "getTimeRangeSchedule: get class's duration schedules failed", log.Err(err))
		return nil, err
	}

	scheduleMap := make(map[entity.TimeRange][]string)
	for i := range schedules {
		for j := range durations {
			if durations[j].MustContain(ctx, schedules[i].StartAt) {
				scheduleMap[durations[j]] = append(scheduleMap[durations[j]], schedules[i].ID)
			}
		}
	}
	return scheduleMap, nil
}

func (c ClassesAssignmentsModel) getScheduleRatios(ctx context.Context, scheduleIDs []string) (map[string]float32, error) {
	panic("implement me")
}

func (c ClassesAssignmentsModel) GetStatistic(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentsViewRequest) ([]*entity.ClassesAssignmentsView, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationIDs:  entity.NullStrings{Strings: request.ClassIDs, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterClass), Valid: true},
	})

	scheduleClassMap := make(map[string]string)
	for i := range relations {
		if _, ok := scheduleClassMap[relations[i].ScheduleID]; !ok {
			scheduleClassMap[relations[i].ScheduleID] = relations[i].RelationID
		}
	}

	rangeSchedule, err := c.getTimeRangeSchedule(ctx, relations, request.Durations, request.Type)
	if err != nil {
		log.Error(ctx, "GetStatistic: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	scheduleIDs := make([]string, 0)
	for _, v := range rangeSchedule {
		for i := range v {
			scheduleIDs = append(scheduleIDs, v[i])
		}
	}

	scheduleRatios, err := c.getScheduleRatios(ctx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "GetStatistic: getScheduleRatios failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	result := make([]*entity.ClassesAssignmentsView, len(request.ClassIDs))
	for i := range request.ClassIDs {
		view := &entity.ClassesAssignmentsView{
			ClassID:        request.ClassIDs[i],
			DurationsRatio: make([]entity.ClassesAssignmentsDurationRatio, len(request.Durations)),
		}
		ids := make([]string, 0)
		for j, durationRatio := range view.DurationsRatio {
			var rationSum float32
			count := 0
			for _, id := range rangeSchedule[entity.TimeRange(durationRatio.Key)] {
				if scheduleClassMap[id] == view.ClassID {
					ids = append(ids, id)
					rationSum += scheduleRatios[id]
					count++
				}
			}
			view.DurationsRatio[j].Ratio = rationSum / float32(count)
		}
		view.Total = len(utils.SliceDeduplication(ids))
		result[i] = view
	}

	return result, nil
}

func (c ClassesAssignmentsModel) getUnattendedMap(ctx context.Context, unattended []*entity.ClassesAssignmentsRecords) (map[string]map[string]bool, error) {
	result := make(map[string]map[string]bool)
	for _, record := range unattended {
		if _, ok := result[record.AttendanceID]; !ok {
			result[record.AttendanceID] = make(map[string]bool)
		}
		result[record.AttendanceID][record.ScheduleID] = true
	}
	return result, nil
}

func (c ClassesAssignmentsModel) GetUnattended(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentsUnattendedViewRequest) ([]*entity.ClassesAssignmentsUnattendedStudentsView, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationID:   sql.NullString{String: request.ClassID, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterClass), Valid: true},
	})

	scheduleClassMap := make(map[string]string)
	for i := range relations {
		if _, ok := scheduleClassMap[relations[i].ScheduleID]; !ok {
			scheduleClassMap[relations[i].ScheduleID] = relations[i].RelationID
		}
	}

	rangeSchedule, err := c.getTimeRangeSchedule(ctx, relations, request.Durations, request.Type)
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	scheduleIDs := make([]string, 0)
	for _, v := range rangeSchedule {
		scheduleIDs = append(scheduleIDs, v...)
	}
	unattended, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx))
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	unattendMap, err := c.getUnattendedMap(ctx, unattended)
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	// get one-page students order by student name
	students := make([]struct {
		ID   string
		Name string
	}, 0)
	result := make([]*entity.ClassesAssignmentsUnattendedStudentsView, 0)
	for i := range students {
		scheduleIDMap := unattendMap[students[i].ID]
		for j := range scheduleIDs {
			view := &entity.ClassesAssignmentsUnattendedStudentsView{
				StudentID:   students[i].ID,
				StudentName: students[i].Name,
			}
			if scheduleIDMap != nil && scheduleIDMap[scheduleIDs[j]] {
				scheduleView := entity.ScheduleView{
					ScheduleID:   scheduleIDs[j],
					ScheduleName: "",
					Type:         request.Type,
				}
				view.Schedule = scheduleView
			}
			result = append(result, view)
		}
	}
	return result, nil
}

func GetClassesAssignmentsModel() IClassesAssignments {
	_classesAssignmentsModelOnce.Do(func() {
		_classesAssignmentsModel = new(ClassesAssignmentsModel)
	})
	return _classesAssignmentsModel
}
