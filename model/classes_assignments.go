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
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IClassesAssignments interface {
	CreateRecord(ctx context.Context, op *entity.Operator, args *entity.AddClassAndLiveAssessmentArgs) (string, error)
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

func (c ClassesAssignmentsModel) CreateRecord(ctx context.Context, op *entity.Operator, data *entity.AddClassAndLiveAssessmentArgs) (string, error) {
	schedule, err := GetScheduleModel().GetPlainByID(ctx, data.ScheduleID)
	if err != nil {
		log.Error(ctx, "CreateRecord: GetPlainByID failed", log.Err(err), log.Any("data", data))
		return "", err
	}
	classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, schedule.ID)
	if err != nil {
		log.Error(ctx, "CreateRecord: GetClassRosterID failed", log.Err(err), log.Any("data", data))
		return "", err
	}
	if classID == "" {
		log.Info(ctx, "CreateRecord: schedule doesn't belong any class", log.Any("data", data))
		return "", nil
	}
	shouldAttendances, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleID:   sql.NullString{String: schedule.ID, Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterStudent), Valid: true},
	})
	if err != nil {
		log.Error(ctx, "CreateRecord: shouldAttendances failed", log.Err(err), log.Any("data", data))
		return "", err
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		existAttendances, err := da.GetClassesAssignmentsDA().QueryTx(ctx, tx, &da.ClassesAssignmentsCondition{
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
				}
				if schedule.ClassType == entity.ScheduleClassTypeHomework {
					insert.ScheduleStartAt = schedule.CreatedAt
				}
				insertRecords = append(insertRecords, insert)
			}
		}
		err = da.GetClassesAssignmentsDA().BatchInsertTx(ctx, tx, insertRecords)
		if err != nil {
			log.Error(ctx, "CreateRecord: BatchInsertTx failed",
				log.Err(err),
				log.Any("data", insertRecords),
				log.Any("should", shouldAttendances))
			return err
		}

		err = da.GetClassesAssignmentsDA().BatchUpdateFinishAndEnd(ctx, tx, schedule.ID, data.AttendanceIDs, data.ClassEndTime)
		if err != nil {
			log.Error(ctx, "CreateRecord: BatchUpdateFinish failed",
				log.Err(err),
				log.Any("data", data))
			return err
		}
		return nil
	})
	return "", err
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

func (c ClassesAssignmentsModel) getScheduleIDMapByType(ctx context.Context, schedules []*entity.Schedule, durations []entity.TimeRange) map[entity.ScheduleInReportType][]string {
	results := make(map[entity.ScheduleInReportType][]string)
	for _, schedule := range schedules {
		for i := range durations {
			if schedule.ClassType == entity.ScheduleClassTypeOnlineClass && durations[i].MustContain(ctx, schedule.StartAt) {
				results[entity.LiveType] = append(results[entity.LiveType], schedule.ID)
			}
			if schedule.ClassType == entity.ScheduleClassTypeHomework && !schedule.IsHomeFun {
				results[entity.StudyType] = append(results[entity.StudyType], schedule.ID)
			}
			if schedule.ClassType == entity.ScheduleClassTypeHomework && schedule.IsHomeFun {
				results[entity.HomeFunType] = append(results[entity.HomeFunType], schedule.ID)
			}
		}
	}
	return results
}

func (c ClassesAssignmentsModel) GetOverview(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentOverViewRequest) ([]*entity.ClassesAssignmentOverView, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationIDs:  entity.NullStrings{Strings: request.ClassIDs.Slice(), Valid: true},
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

	//min, max, err := c.getMinAndMax(ctx, request.Durations)
	//if err != nil {
	//	log.Error(ctx, "GetOverview: extract time duration failed", log.Err(err), log.Any("request", request))
	//	return nil, err
	//}

	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs: entity.NullStrings{Strings: scheduleIDs, Valid: true},
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
	scheduleTypeMaps := c.getScheduleIDMapByType(ctx, schedules, request.Durations.TimeRangeSlice())
	if scheduleTypeMaps[entity.LiveType] != nil {
		overviews[0].Count = len(utils.SliceDeduplication(scheduleTypeMaps[entity.LiveType]))
	}
	if scheduleTypeMaps[entity.StudyType] != nil {
		overviews[1].Count = len(utils.SliceDeduplication(scheduleTypeMaps[entity.StudyType]))
	}
	if scheduleTypeMaps[entity.HomeFunType] != nil {
		overviews[2].Count = len(utils.SliceDeduplication(scheduleTypeMaps[entity.HomeFunType]))
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

func (c ClassesAssignmentsModel) getScheduleRatios(ctx context.Context, scheduleIDs []string) (map[string][]int, error) {
	records, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx), &da.ClassesAssignmentsCondition{
		ScheduleIDs: entity.NullStrings{Strings: scheduleIDs, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "getScheduleRatios: query failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs))
		return nil, err
	}
	shouldActualMap := make(map[string][]int)
	for _, record := range records {
		if shouldActualMap[record.ScheduleID] == nil {
			shouldActualMap[record.ScheduleID] = make([]int, 2)
		}
		shouldActualMap[record.ScheduleID][0]++
		if record.FinishCount > 0 {
			shouldActualMap[record.ScheduleID][1]++
		}
	}
	return shouldActualMap, nil
}

func (c ClassesAssignmentsModel) GetStatistic(ctx context.Context, op *entity.Operator, request *entity.ClassesAssignmentsViewRequest) ([]*entity.ClassesAssignmentsView, error) {
	relations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		RelationIDs:  entity.NullStrings{Strings: request.ClassIDs.Slice(), Valid: true},
		RelationType: sql.NullString{String: string(entity.ScheduleRelationTypeClassRosterClass), Valid: true},
	})

	scheduleClassMap := make(map[string]string)
	for i := range relations {
		if _, ok := scheduleClassMap[relations[i].ScheduleID]; !ok {
			scheduleClassMap[relations[i].ScheduleID] = relations[i].RelationID
		}
	}

	filterSchedules, scheduleIDRangeIDMap, err := c.getScheduleIDMapByTimeRange(ctx, relations, request.Durations.TimeRangeSlice(), request.Type)
	if err != nil {
		log.Error(ctx, "GetStatistic: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	scheduleIDs := make([]string, len(filterSchedules))
	for i := range filterSchedules {
		scheduleIDs[i] = filterSchedules[i].ID
	}

	scheduleShouldActualMap, err := c.getScheduleRatios(ctx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "GetStatistic: getScheduleRatios failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	log.Debug(ctx, "get schedule ratios successfully", log.Any("ratio", scheduleShouldActualMap))

	result := make([]*entity.ClassesAssignmentsView, len(request.ClassIDs.Slice()))
	for i, classID := range request.ClassIDs.Slice() {
		view := &entity.ClassesAssignmentsView{
			ClassID:        classID,
			DurationsRatio: make([]entity.ClassesAssignmentsDurationRatio, len(request.Durations.Slice())),
		}
		ids := make([]string, 0)
		for j, duration := range request.Durations.TimeRangeSlice() {
			var rationSum float32
			count := 0
			for _, id := range scheduleIDRangeIDMap[duration] {
				if scheduleClassMap[id] == view.ClassID {
					ids = append(ids, id)
					if scheduleShouldActualMap[id] != nil && scheduleShouldActualMap[id][0] != 0 {
						rationSum += float32(scheduleShouldActualMap[id][1]) / float32(scheduleShouldActualMap[id][0])
					}
					count++
				}
			}
			view.DurationsRatio[j].Key = string(duration)
			if count != 0 {
				view.DurationsRatio[j].Ratio = rationSum / float32(count)
			}

		}
		view.Total = len(utils.SliceDeduplication(ids))
		result[i] = view
	}

	log.Debug(ctx, "GetStatistic successfully", log.Any("result", result))
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

	filterSchedules, _, err := c.getScheduleIDMapByTimeRange(ctx, relations, request.Durations.TimeRangeSlice(), request.Type)
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	scheduleIDs := make([]string, len(filterSchedules))
	scheduleIDNameMap := make(map[string]string)
	for i := range filterSchedules {
		scheduleIDs[i] = filterSchedules[i].ID
		scheduleIDNameMap[filterSchedules[i].ID] = filterSchedules[i].Title
	}

	unattended, err := da.GetClassesAssignmentsDA().QueryTx(ctx, dbo.MustGetDB(ctx), &da.ClassesAssignmentsCondition{
		ScheduleIDs:     entity.NullStrings{Strings: scheduleIDs, Valid: true},
		FinishCountsLTE: sql.NullInt64{Int64: 0, Valid: true},
	})
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}
	unattendedMap, err := c.getUnattendedMap(ctx, unattended)
	if err != nil {
		log.Error(ctx, "GetUnattended: extract time duration failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	// TODO: get one-page students order by student name
	// Now, give all unattendedMap
	type studentType struct {
		ID   string
		Name string
	}
	students := make([]*studentType, len(unattendedMap))
	index := 0
	for k, _ := range unattendedMap {
		students[index] = &studentType{ID: k}
		index++
	}

	result := make([]*entity.ClassesAssignmentsUnattendedStudentsView, 0)
	for i := range students {
		scheduleIDMap := unattendedMap[students[i].ID]
		for j := range scheduleIDs {
			view := &entity.ClassesAssignmentsUnattendedStudentsView{
				StudentID:   students[i].ID,
				StudentName: students[i].Name,
			}
			if scheduleIDMap != nil && scheduleIDMap[scheduleIDs[j]] {
				scheduleView := entity.ScheduleView{
					ScheduleID:   scheduleIDs[j],
					ScheduleName: scheduleIDNameMap[scheduleIDs[j]],
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
