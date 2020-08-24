package dyschedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/daschedule"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/dyschedule"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils/dynamodbhelper"
	"sync"
)

type IScheduleModel interface {
	Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error)
	Update(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewdata *entity.ScheduleUpdateView) (string, error)
	Delete(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error
	Query(ctx context.Context, tx *dbo.DBContext, condition *daschedule.ScheduleCondition) ([]*entity.ScheduleListView, error)
	Page(ctx context.Context, tx *dbo.DBContext, condition *daschedule.ScheduleCondition) (int, []*entity.ScheduleSeachView, error)
	//PageByTeacherID(ctx context.Context, tx *dbo.DBContext, condition *dyschedule.ScheduleCondition) (string, []*entity.ScheduleSeachView, error)
	GetByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.ScheduleDetailsView, error)
}
type scheduleModel struct {
	testScheduleRepeatFlag bool
}

func (s *scheduleModel) Page(ctx context.Context, tx *dbo.DBContext, condition *daschedule.ScheduleCondition) (int, []*entity.ScheduleSeachView, error) {
	var scheduleList []*entity.Schedule
	total, err := daschedule.GetScheduleTeacherDA().Page(ctx, condition, &scheduleList)
	if err != nil {
		return 0, nil, err
	}
	var result = make([]*entity.ScheduleSeachView, len(scheduleList))
	for i, item := range scheduleList {
		baseinfo, err := s.getBasicInfo(ctx, tx, item)
		if err != nil {
			return 0, nil, err
		}
		result[i] = &entity.ScheduleSeachView{
			ID:            item.ID,
			StartAt:       item.StartAt,
			EndAt:         item.EndAt,
			ScheduleBasic: *baseinfo,
		}
	}
	return total, result, nil
}

func (s *scheduleModel) addRepeatSchedule(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule) (string, error) {
	scheduleList, err := s.RepeatSchedule(ctx, schedule)
	if err != nil {
		log.Error(ctx, "daschedule repeat error", log.Err(err), log.Any("daschedule", schedule))
		return "", err
	}
	teacherSchedules := make([]*entity.TeacherSchedule, len(scheduleList)*len(schedule.TeacherIDs))
	index := 0
	for _, item := range scheduleList {
		item.ID = utils.NewID()
		for _, teacherID := range item.TeacherIDs {
			tsItem := &entity.TeacherSchedule{
				TeacherID:  teacherID,
				ScheduleID: schedule.ID,
				StartAt:    schedule.StartAt,
			}
			teacherSchedules[index] = tsItem
			index++
		}
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// add to schedules
		_, err = daschedule.GetScheduleDA().BatchInsert(ctx, tx, scheduleList)
		if err != nil {
			log.Error(ctx, "schedule batchInsert error", log.Err(err))
			return err
		}

		// add to teachers_schedules
		_, err = daschedule.GetScheduleTeacherDA().BatchInsert(ctx, tx, teacherSchedules)
		if err != nil {
			log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("teacherSchedules", teacherSchedules))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(scheduleList) > 0 {
		return scheduleList[0].ID, errors.New("")
	}
	return "", errors.New("")
}
func (s *scheduleModel) Add(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewdata *entity.ScheduleAddView) (string, error) {
	schedule := viewdata.Convert()
	schedule.CreatedID = op.UserID
	if viewdata.ModeType == entity.ModeTypeRepeat {
		return s.addRepeatSchedule(ctx, tx, schedule)
	} else {
		schedule.ID = utils.NewID()
		err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			_, err := daschedule.GetScheduleDA().InsertTx(ctx, tx, schedule)
			if err != nil {
				return err
			}
			teacherSchedules := make([]*entity.TeacherSchedule, len(schedule.TeacherIDs))
			for i, item := range viewdata.TeacherIDs {
				teacherSchedule := &entity.TeacherSchedule{
					ID:         utils.NewID(),
					TeacherID:  item,
					ScheduleID: schedule.ID,
					DeletedAt:  0,
				}
				teacherSchedules[i] = teacherSchedule
			}
			// add to teachers_schedules
			_, err = daschedule.GetScheduleTeacherDA().BatchInsert(ctx, tx, teacherSchedules)
			if err != nil {
				log.Error(ctx, "teachers_schedules batchInsert error", log.Err(err), log.Any("teacherSchedules", teacherSchedules))
				return err
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		return schedule.ID, nil
	}
}

func (s *scheduleModel) Update(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, viewdata *entity.ScheduleUpdateView) (string, error) {
	// TODO: check permission
	if !viewdata.EditType.Valid() {
		err := errors.New("update daschedule: invalid type")
		log.Error(ctx, err.Error(), log.String("edit_type", string(viewdata.EditType)))
		return "", err
	}
	if _, err := dyschedule.GetScheduleDA().GetByID(ctx, viewdata.ID); err != nil {
		if err == constant.ErrRecordNotFound {
			log.Error(ctx, "update daschedule: record not found", log.Err(err))
		} else {
			log.Error(ctx, "update daschedule: get daschedule by id failed", log.String("id", viewdata.ID))
		}
		return "", err
	}
	if err := s.Delete(ctx, tx, op, viewdata.ID, viewdata.EditType); err != nil {
		log.Error(ctx, "update daschedule: delete failed",
			log.Err(err),
			log.String("id", viewdata.ID),
			log.String("edit_type", string(viewdata.EditType)),
		)
		return "", err
	}
	id, err := s.Add(ctx, tx, op, &viewdata.ScheduleAddView)
	if err != nil {
		log.Error(ctx, "update daschedule: delete failed",
			log.Err(err),
			log.Any("schedule_add_view", viewdata.ScheduleAddView),
		)
		return "", err
	}
	return id, nil
}

func (s *scheduleModel) Delete(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, id string, editType entity.ScheduleEditType) error {
	// TODO: check permission
	schedule, err := dyschedule.GetScheduleDA().GetByID(ctx, id)
	if err != nil {
		if err == constant.ErrRecordNotFound {
			log.Warn(ctx, "delete daschedule: record not found", log.String("id", id))
			return nil
		}
		log.Error(ctx, "delete daschedule: get daschedule by id failed",
			log.String("id", id))
		return err
	}
	var deletingTeacherSchedulePKs [][2]string
	switch editType {
	case entity.ScheduleEditOnlyCurrent:
		if err := dyschedule.GetScheduleDA().Delete(ctx, id); err != nil {
			log.Error(ctx, "delete daschedule: delete failed",
				log.String("id", id), log.String("edit_type", string(editType)))
			return err
		}
		for _, teacherID := range schedule.TeacherIDs {
			deletingTeacherSchedulePKs = append(deletingTeacherSchedulePKs, [2]string{teacherID, id})
		}
	case entity.ScheduleEditWithFollowing:
		cond := dyschedule.ScheduleCondition{
			RepeatID: schedule.RepeatID,
			StartAt:  schedule.StartAt,
		}
		cond.Init(constant.GSI_Schedule_RepeatIDAndStartAt, dynamodbhelper.SortKeyGreaterThanEqual)
		schedules, err := dyschedule.GetScheduleDA().Query(ctx, &cond)
		if err != nil {
			log.Error(ctx, "delete daschedule: query failed", log.Any("cond", cond))
			return err
		}
		var ids []string
		for _, schedule := range schedules {
			ids = append(ids, schedule.ID)
			for _, teacherID := range schedule.TeacherIDs {
				deletingTeacherSchedulePKs = append(deletingTeacherSchedulePKs, [2]string{teacherID, id})
			}
		}
		if err = dyschedule.GetScheduleDA().BatchDelete(ctx, ids); err != nil {
			log.Error(ctx, "delete daschedule: batch delete failed", log.Err(err))
			return err
		}
	default:
		err := fmt.Errorf("delete daschedule: invalid edit type")
		log.Error(ctx, err.Error(), log.String("edit_type", string(editType)))
		return err
	}
	if len(deletingTeacherSchedulePKs) > 0 {
		if err := dyschedule.GetTeacherScheduleDA().BatchDelete(ctx, deletingTeacherSchedulePKs); err != nil {
			log.Error(ctx, "delete daschedule: batch delete teacher_schedule failed",
				log.Any("pks", deletingTeacherSchedulePKs))
			return err
		}
	}
	return nil
}

func (s *scheduleModel) PageByTeacherID(ctx context.Context, tx *dbo.DBContext, condition *dyschedule.ScheduleCondition) (string, []*entity.ScheduleSeachView, error) {
	tsCondition := dyschedule.TeacherScheduleCondition{
		TeacherID: condition.TeacherID,
		StartAt:   condition.StartAt,
	}
	tsCondition.Init(constant.GSI_TeacherSchedule_TeacherAndStartAt, dynamodbhelper.SortKeyGreaterThanEqual)

	lastKey, data, err := dyschedule.GetTeacherScheduleDA().Page(ctx, tsCondition)
	ids := make([]string, len(data))
	for i, item := range data {
		ids[i] = item.ScheduleID
	}
	scheduleList, err := dyschedule.GetScheduleDA().BatchGetByIDs(ctx, ids)
	if err != nil {
		log.Error(ctx, "PageByTeacherID:batch get by daschedule ids error", log.Err(err), log.Strings("ids", ids))
		return "", nil, err
	}

	result := make([]*entity.ScheduleSeachView, len(scheduleList))
	for i, item := range scheduleList {
		viewdata := &entity.ScheduleSeachView{
			ID:      item.ID,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		}
		basicInfo, err := s.getBasicInfo(ctx, tx, item)
		if err != nil {
			log.Error(ctx, "PageByTeacherID:getBasicInfo error", log.Err(err), log.Any("scheduleItem", item))
			return "", nil, err
		}
		viewdata.ScheduleBasic = *basicInfo
		result[i] = viewdata
	}

	return lastKey, result, nil
}

func (s *scheduleModel) Query(ctx context.Context, tx *dbo.DBContext, condition *daschedule.ScheduleCondition) ([]*entity.ScheduleListView, error) {
	var scheduleList []*entity.Schedule
	err := daschedule.GetScheduleDA().Query(ctx, condition, &scheduleList)
	if err != nil {
		log.Error(ctx, "daschedule query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.ScheduleListView, len(scheduleList))
	for i, item := range scheduleList {
		result[i] = &entity.ScheduleListView{
			ID:      item.ID,
			Title:   item.Title,
			StartAt: item.StartAt,
			EndAt:   item.EndAt,
		}
	}
	return result, nil
}

func (s *scheduleModel) getBasicInfo(ctx context.Context, tx *dbo.DBContext, schedule *entity.Schedule) (*entity.ScheduleBasic, error) {
	result := &entity.ScheduleBasic{}
	if schedule.ClassID != "" {
		classService, err := external.GetClassServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		classInfos, err := classService.BatchGet(ctx, []string{schedule.ClassID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetClassServiceProvider BatchGet error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		if len(classInfos) > 0 {
			result.Class = entity.ShortInfo{
				ID:   classInfos[0].ID,
				Name: classInfos[0].Name,
			}
		}
	}

	if schedule.SubjectID != "" {
		subjectService, err := external.GetSubjectServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		subjectInfos, err := subjectService.BatchGet(ctx, []string{schedule.SubjectID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetSubjectServiceProvider BatchGet error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		if len(subjectInfos) > 0 {
			result.Subject = entity.ShortInfo{
				ID:   subjectInfos[0].ID,
				Name: subjectInfos[0].Name,
			}
		}
	}
	if schedule.ProgramID != "" {
		programService, err := external.GetProgramServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		programInfos, err := programService.BatchGet(ctx, []string{schedule.ProgramID})
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetProgramServiceProvider BatchGet error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		if len(programInfos) > 0 {
			result.Program = entity.ShortInfo{
				ID:   programInfos[0].ID,
				Name: programInfos[0].Name,
			}
		}
	}
	var scheduleTeacherList []*entity.TeacherSchedule
	err := daschedule.GetScheduleTeacherDA().Query(ctx, &daschedule.ScheduleTeacherCondition{
		ScheduleID: sql.NullString{
			String: schedule.ID,
			Valid:  true,
		},
	}, &scheduleTeacherList)
	if err != nil {
		return nil, err
	}
	if len(scheduleTeacherList) != 0 {
		teacherIDs := make([]string, len(scheduleTeacherList))
		for _, item := range scheduleTeacherList {
			teacherIDs = append(teacherIDs, item.TeacherID)
		}
		result.Teachers = make([]entity.ShortInfo, len(teacherIDs))
		teacherService, err := external.GetTeacherServiceProvider()
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		teacherInfos, err := teacherService.BatchGet(ctx, teacherIDs)
		if err != nil {
			log.Error(ctx, "getBasicInfo:GetTeacherServiceProvider BatchGet error", log.Err(err), log.Any("daschedule", schedule))
			return nil, err
		}
		for i, item := range teacherInfos {
			result.Teachers[i] = entity.ShortInfo{
				ID:   item.ID,
				Name: item.Name,
			}
		}
	}
	// TODO LessonPlan Attachment

	return result, nil
}

func (s *scheduleModel) GetByID(ctx context.Context, tx *dbo.DBContext, id string) (*entity.ScheduleDetailsView, error) {
	var schedule *entity.Schedule
	err := daschedule.GetScheduleDA().Get(ctx, id, schedule)
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}

	result := &entity.ScheduleDetailsView{
		ID:          schedule.ID,
		Title:       schedule.Title,
		OrgID:       schedule.OrgID,
		StartAt:     schedule.StartAt,
		EndAt:       schedule.EndAt,
		ModeType:    schedule.ModeType,
		ClassType:   schedule.ClassType,
		DueAt:       schedule.DueAt,
		Description: schedule.Description,
		Version:     schedule.Version,
		RepeatID:    schedule.RepeatID,
		Repeat:      schedule.Repeat,
	}
	basicInfo, err := s.getBasicInfo(ctx, tx, schedule)
	if err != nil {
		log.Error(ctx, "getBasicInfo error", log.Err(err))
		return nil, err
	}
	result.ScheduleBasic = *basicInfo
	return result, nil
}

var (
	_scheduleOnce  sync.Once
	_scheduleModel IScheduleModel
)

func GetScheduleModel() IScheduleModel {
	_scheduleOnce.Do(func() {
		_scheduleModel = &scheduleModel{}
	})
	return _scheduleModel
}
