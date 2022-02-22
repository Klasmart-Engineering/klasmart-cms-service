package model

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IClassesAssignments interface {
	CreateRecord(ctx context.Context, op *entity.Operator, data *v2.ScheduleEndClassCallBackReq) error
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

func (c ClassesAssignmentsModel) CreateRecord(ctx context.Context, op *entity.Operator, data *v2.ScheduleEndClassCallBackReq) error {
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
	shouldAttendances, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleID:   sql.NullString{String: schedule.ID, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterStudent), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "CreateRecord: shouldAttendances failed", log.Err(err), log.Any("data", data))
		return err
	}

	existAttendances, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx), &da.ClassesAssignmentsCondition{
		ClassID:    sql.NullString{String: classID, Valid: true},
		ScheduleID: sql.NullString{String: schedule.ID, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "CreateRecord: get exists failed",
			log.Err(err),
			log.Any("data", data),
			log.Any("should", shouldAttendances))
		return err
	}

	insertRecords := make([]*entity.ClassesAssignmentsRecords, 0)
	for i := range shouldAttendances {
		exist := false
		for j := range existAttendances {
			if shouldAttendances[i].RelationID == existAttendances[j].AttendanceID {
				exist = true
				break
			}
		}
		if !exist {
			insert := &entity.ClassesAssignmentsRecords{
				ID:              utils.NewID(),
				ClassID:         classID,
				ScheduleID:      schedule.ID,
				AttendanceID:    shouldAttendances[i].RelationID,
				ScheduleType:    entity.NewScheduleInReportType(schedule.ClassType, schedule.IsHomeFun),
				ScheduleStartAt: schedule.StartAt,
				CreateAt:        time.Now().Unix(),
				LastEndAt:       data.ClassEndAt,
			}
			if schedule.ClassType == entity.ScheduleClassTypeHomework {
				insert.ScheduleStartAt = schedule.CreatedAt
			}
			insertRecords = append(insertRecords, insert)
		}
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err = da.GetClassesAssignmentsDA().BatchInsertTx(ctx, tx, insertRecords)
		if err != nil {
			log.Error(ctx, "CreateRecord: BatchInsertTx failed",
				log.Err(err),
				log.Any("data", insertRecords),
				log.Any("should", shouldAttendances))
			return err
		}

		err = da.GetClassesAssignmentsDA().BatchUpdateFinishAndEnd(ctx, tx, schedule.ID, data.AttendanceIDs, data.ClassEndAt)
		if err != nil {
			log.Error(ctx, "CreateRecord: BatchUpdateFinish failed",
				log.Err(err),
				log.Any("data", data))
			return err
		}
		return nil
	})
	return err
}

func (c ClassesAssignmentsModel) getMinAndMax(ctx context.Context, timeRanges []entity.TimeRange) (int64, int64, error) {
	min := int64(^uint64(0) >> 1)
	max := int64(0)
	for i := range timeRanges {
		startAt, endAt, err := timeRanges[i].Value(ctx)
		if err != nil {
			log.Error(ctx, "getMinAndMax: extract time duration failed", log.Err(err), log.Any("timeRanges", timeRanges))
			return 0, 0, err
		}
		if min > startAt {
			min = startAt
		}
		if max < endAt {
			max = endAt
		}
	}
	return min, max, nil
}

func (c ClassesAssignmentsModel) getScheduleIDMapByType(ctx context.Context, schedules []*entity.Schedule, durations []entity.TimeRange) ([]string, map[entity.ScheduleInReportType][]string) {
	results := make(map[entity.ScheduleInReportType][]string)
	for _, schedule := range schedules {
		for i := range durations {
			if schedule.ClassType == entity.ScheduleClassTypeOnlineClass && durations[i].MustContain(ctx, schedule.StartAt) {
				results[entity.LiveType] = append(results[entity.LiveType], schedule.ID)
			}
			if schedule.ClassType == entity.ScheduleClassTypeHomework && !schedule.IsHomeFun && durations[i].MustContain(ctx, schedule.CreatedAt) {
				results[entity.StudyType] = append(results[entity.StudyType], schedule.ID)
			}
			if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun && durations[i].MustContain(ctx, schedule.CreatedAt) {
				results[entity.HomeFunType] = append(results[entity.HomeFunType], schedule.ID)
			}
		}
	}

	scheduleIDs := make([]string, 0)
	for k, v := range results {
		results[k] = utils.SliceDeduplicationExcludeEmpty(v)
		scheduleIDs = append(scheduleIDs, results[k]...)
	}

	return scheduleIDs, results
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

	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs: entity.NullStrings{Strings: scheduleIDs, Valid: true},
		ClassTypes: entity.NullStrings{Strings: []string{
			string(entity.ScheduleClassTypeOnlineClass),
			string(entity.ScheduleClassTypeHomework),
		}, Valid: true},
		//StartAtGe: sql.NullInt64{Int64: min, Valid: true},
		//StartAtLt: sql.NullInt64{Int64: max, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetOverview: get class's duration schedules failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	overviews := []*entity.ClassesAssignmentOverView{
		{Type: string(entity.LiveType), Count: 0},
		{Type: string(entity.StudyType), Count: 0},
		{Type: string(entity.HomeFunType), Count: 0},
	}
	durationScheduleIDs, scheduleTypeMaps := c.getScheduleIDMapByType(ctx, schedules, request.Durations)
	shouldAndActual, err := c.countScheduleShouldAndActual(ctx, op, durationScheduleIDs)
	if err != nil {
		log.Error(ctx, "GetOverView: get ratios failed", log.Err(err), log.Any("request", request), log.Strings("schedule_ids", durationScheduleIDs))
		return nil, err
	}

	if scheduleTypeMaps[entity.LiveType] != nil {
		overviews[0].Count = len(scheduleTypeMaps[entity.LiveType])
		overviews[0].Ratio = c.getScheduleTotalRatios(ctx, scheduleTypeMaps[entity.LiveType], shouldAndActual)
	}
	if scheduleTypeMaps[entity.StudyType] != nil {
		overviews[1].Count = len(scheduleTypeMaps[entity.StudyType])
		overviews[1].Ratio = c.getScheduleTotalRatios(ctx, scheduleTypeMaps[entity.StudyType], shouldAndActual)
	}
	if scheduleTypeMaps[entity.HomeFunType] != nil {
		overviews[2].Count = len(scheduleTypeMaps[entity.HomeFunType])
		overviews[2].Ratio = c.getScheduleTotalRatios(ctx, scheduleTypeMaps[entity.HomeFunType], shouldAndActual)
	}
	return overviews, nil
}

func (c ClassesAssignmentsModel) getScheduleIDMapByTimeRange(ctx context.Context, relations []*entity.ScheduleRelation, durations []entity.TimeRange, kind string) ([]*entity.Schedule, map[entity.TimeRange][]string, error) {
	ids := make([]string, len(relations))
	for i := range relations {
		ids[i] = relations[i].ScheduleID
	}
	min, max, err := c.getMinAndMax(ctx, durations)
	if err != nil {
		log.Error(ctx, "getScheduleIDMapByTimeRange: extract time duration failed", log.Err(err))
		return nil, nil, err
	}

	condition := &entity.ScheduleQueryCondition{IDs: entity.NullStrings{Strings: ids, Valid: true}}

	if kind == string(entity.LiveType) {
		classType := string(entity.ScheduleClassTypeOnlineClass)
		condition.ClassTypes = entity.NullStrings{Strings: []string{classType}, Valid: true}
		condition.StartAtGe = sql.NullInt64{Int64: min, Valid: true}
		condition.StartAtLt = sql.NullInt64{Int64: max, Valid: true}
	}
	if kind == string(entity.StudyType) {
		classType := string(entity.ScheduleClassTypeHomework)
		condition.ClassTypes = entity.NullStrings{Strings: []string{classType}, Valid: true}
		condition.CreateAtGe = sql.NullInt64{Int64: min, Valid: true}
		condition.CreateAtLt = sql.NullInt64{Int64: max, Valid: true}
		condition.IsHomefun = sql.NullBool{Bool: false, Valid: true}
	}
	if kind == string(entity.HomeFunType) {
		classType := string(entity.ScheduleClassTypeHomework)
		condition.ClassTypes = entity.NullStrings{Strings: []string{classType}, Valid: true}
		condition.CreateAtGe = sql.NullInt64{Int64: min, Valid: true}
		condition.CreateAtLt = sql.NullInt64{Int64: max, Valid: true}
		condition.IsHomefun = sql.NullBool{Bool: true, Valid: true}
	}
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, condition)
	if err != nil {
		log.Error(ctx, "getScheduleIDMapByTimeRange: get class's duration schedules failed", log.Err(err))
		return nil, nil, err
	}

	scheduleMap := make(map[entity.TimeRange][]string)
	filterSchedules := make([]*entity.Schedule, 0)
	for i := range schedules {
		startAt := schedules[i].StartAt
		if schedules[i].ClassType == entity.ScheduleClassTypeHomework {
			startAt = schedules[i].CreatedAt
		}
		for j := range durations {
			if durations[j].MustContain(ctx, startAt) {
				scheduleMap[durations[j]] = append(scheduleMap[durations[j]], schedules[i].ID)
				filterSchedules = append(filterSchedules, schedules[i])
			}
		}
	}
	return filterSchedules, scheduleMap, nil
}

func (c ClassesAssignmentsModel) countScheduleShouldAndActual(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]int, error) {
	allStudents, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleIDs:  entity.NullStrings{Strings: scheduleIDs, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterStudent), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "countScheduleShouldAndActual: query relation failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	allMap := make(map[string][]string)
	for _, relation := range allStudents {
		allMap[relation.ScheduleID] = append(allMap[relation.ScheduleID], relation.RelationID)
	}

	attendance, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx), &da.ClassesAssignmentsCondition{
		ScheduleIDs:    entity.NullStrings{Strings: scheduleIDs, Valid: true},
		FinishCountsGT: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "countScheduleShouldAndActual: query attendance failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	actualMap := make(map[string][]string)
	for _, r := range attendance {
		actualMap[r.ScheduleID] = append(actualMap[r.ScheduleID], r.AttendanceID)
	}

	shouldActualMap := make(map[string][]int, len(allMap))
	for k, v := range allMap {
		shouldActualMap[k] = make([]int, 2)
		shouldActualMap[k][0] = len(utils.SliceDeduplication(v))
		shouldActualMap[k][1] = len(utils.SliceDeduplication(actualMap[k]))
	}
	return shouldActualMap, nil
}

func (c ClassesAssignmentsModel) getScheduleTotalRatios(ctx context.Context, scheduleIDs []string, shouldActualMap map[string][]int) float32 {
	if len(scheduleIDs) <= 0 {
		return 0
	}

	sumRatios := float32(0)
	for _, id := range scheduleIDs {
		if shouldActualMap != nil && shouldActualMap[id] != nil && shouldActualMap[id][0] != 0 {
			sumRatios += float32(shouldActualMap[id][1]) / float32(shouldActualMap[id][0])
		}
	}
	return sumRatios / float32(len(scheduleIDs))
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

	filterSchedules, scheduleIDRangeIDMap, err := c.getScheduleIDMapByTimeRange(ctx, relations, request.Durations, request.Type)
	if err != nil {
		log.Error(ctx, "GetStatistic: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	scheduleIDs := make([]string, len(filterSchedules))
	for i := range filterSchedules {
		scheduleIDs[i] = filterSchedules[i].ID
	}

	scheduleShouldActualMap, err := c.countScheduleShouldAndActual(ctx, op, scheduleIDs)
	if err != nil {
		log.Error(ctx, "GetStatistic: countScheduleShouldAndActual failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	log.Debug(ctx, "get schedule ratios successfully", log.Any("ratio", scheduleShouldActualMap))

	result := make([]*entity.ClassesAssignmentsView, len(request.ClassIDs))
	for i, classID := range request.ClassIDs {
		view := &entity.ClassesAssignmentsView{
			ClassID:        classID,
			DurationsRatio: make([]entity.ClassesAssignmentsDurationRatio, len(request.Durations)),
		}
		ids := make([]string, 0)
		for j, duration := range request.Durations {
			var rationSum float32
			count := 0
			for _, id := range scheduleIDRangeIDMap[duration] {
				if scheduleClassMap[id] == view.ClassID {
					ids = append(ids, id)
					if scheduleShouldActualMap[id] != nil && scheduleShouldActualMap[id][0] != 0 {
						rationSum += float32(scheduleShouldActualMap[id][1]) / float32(scheduleShouldActualMap[id][0])
						// ignore the attendees as 0
						count++
					}
					log.Debug(ctx, "ratio_sum",
						log.String("class_id", view.ClassID),
						log.String("duration", string(duration)),
						log.String("schedule_id", id),
						log.Any("should_actual", scheduleShouldActualMap[id]))
				}
			}
			view.DurationsRatio[j].Key = string(duration)
			log.Debug(ctx, "ratio_denominator",
				log.String("class_id", view.ClassID),
				log.String("duration", string(duration)),
				log.Int("count", count))
			if count != 0 {
				view.DurationsRatio[j].Ratio = rationSum / float32(count)
			}

		}
		log.Debug(ctx, "ratio_schedules",
			log.String("class_id", view.ClassID),
			log.Strings("schedule_ids", utils.SliceDeduplication(ids)))
		view.Total = len(utils.SliceDeduplication(ids))
		result[i] = view
	}

	log.Debug(ctx, "GetStatistic successfully", log.Any("result", result))
	return result, nil
}

func (c ClassesAssignmentsModel) getShouldAttendedSchedulesMap(ctx context.Context, op *entity.Operator, scheduleIDs []string, studentIDs []string) (map[string][]string, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleIDs:  entity.NullStrings{Strings: scheduleIDs, Valid: true},
		RelationIDs:  entity.NullStrings{Strings: studentIDs, Valid: studentIDs != nil},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterStudent), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetAllAttendanceSchedulesMap: query failed",
			log.Err(err),
			log.Strings("student_ids", studentIDs),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	result := make(map[string][]string)
	for _, relation := range relations {
		result[relation.RelationID] = append(result[relation.RelationID], relation.ScheduleID)
	}
	return result, nil
}

func (c ClassesAssignmentsModel) getActualAttendedSchedulesMap(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string][]string, error) {
	records, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx), &da.ClassesAssignmentsCondition{
		ScheduleIDs:    entity.NullStrings{Strings: scheduleIDs, Valid: true},
		FinishCountsGT: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "getActualAttendedSchedulesMap: query failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	result := make(map[string][]string)
	for _, record := range records {
		result[record.AttendanceID] = append(result[record.AttendanceID], record.ScheduleID)
	}
	return result, nil
}

func (c ClassesAssignmentsModel) getUnattendedSchedulesMap(ctx context.Context, shouldMap map[string][]string, actualMap map[string][]string) map[string][]string {
	result := make(map[string][]string, len(shouldMap))
	for k, v := range shouldMap {
		attendances := actualMap[k]
		for _, attendanceID := range v {
			attended := false
			for i := range attendances {
				if attendanceID == attendances[i] {
					attended = true
					break
				}
			}
			if !attended {
				result[k] = append(result[k], attendanceID)
			}
		}
		log.Debug(ctx, "should and actual",
			log.String("student_id", k),
			log.Any("should", v),
			log.Any("actual", attendances),
			log.Any("unattended", result[k]))
	}
	return result
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

	filterSchedules, _, err := c.getScheduleIDMapByTimeRange(ctx, relations, request.Durations, request.Type)
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	scheduleIDs := make([]string, len(filterSchedules))
	scheduleIDScheduleMap := make(map[string]*entity.Schedule)
	for i := range filterSchedules {
		scheduleIDs[i] = filterSchedules[i].ID
		scheduleIDScheduleMap[filterSchedules[i].ID] = filterSchedules[i]
	}

	shouldAttendedMap, err := c.getShouldAttendedSchedulesMap(ctx, op, scheduleIDs, nil)
	if err != nil {
		log.Error(ctx, "GetUnattended: should attended schedules map",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	actualAttendedMap, err := c.getActualAttendedSchedulesMap(ctx, op, scheduleIDs)
	if err != nil {
		log.Error(ctx, "GetUnattended: actual attended schedules map",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}

	unattendedMap := c.getUnattendedSchedulesMap(ctx, shouldAttendedMap, actualAttendedMap)

	result := make([]*entity.ClassesAssignmentsUnattendedStudentsView, 0)
	for k, v := range unattendedMap {
		for _, scheduleID := range v {
			view := &entity.ClassesAssignmentsUnattendedStudentsView{
				StudentID: k,
				Schedule: entity.ScheduleView{
					ScheduleID:   scheduleID,
					ScheduleName: scheduleIDScheduleMap[scheduleID].Title,
					Type:         request.Type,
				},
			}
			if scheduleIDScheduleMap[scheduleID] != nil && scheduleIDScheduleMap[scheduleID].ClassType == entity.ScheduleClassTypeOnlineClass {
				view.Time = scheduleIDScheduleMap[scheduleID].StartAt
			}
			if scheduleIDScheduleMap[scheduleID] != nil && scheduleIDScheduleMap[scheduleID].ClassType == entity.ScheduleClassTypeHomework {
				view.Time = scheduleIDScheduleMap[scheduleID].CreatedAt
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
