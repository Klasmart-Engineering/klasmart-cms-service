package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IAssessmentModel interface {
	Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error)
	List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error)
	Add(ctx context.Context, cmd entity.AddAssessmentCommand) (string, error)
	Update(ctx context.Context, cmd entity.UpdateAssessmentCommand) error
}

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type assessmentModel struct{}

type outcomeSliceSortByAssumedAndName []*entity.Outcome

func (s outcomeSliceSortByAssumedAndName) Len() int {
	return len(s)
}

func (s outcomeSliceSortByAssumedAndName) Less(i, j int) bool {
	if s[i].Assumed && !s[j].Assumed {
		return true
	} else if !s[i].Assumed && s[j].Assumed {
		return false
	} else {
		return s[i].Name < s[j].Name
	}
}

func (s outcomeSliceSortByAssumedAndName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (a *assessmentModel) Detail(ctx context.Context, tx *dbo.DBContext, id string) (*entity.AssessmentDetailView, error) {
	cacheResult, err := da.GetAssessmentRedisDA().Detail(ctx, id)
	if err != nil {
		log.Info(ctx, "get assessment detail: get failed from cache",
			log.Err(err),
			log.String("id", "id"),
		)
	} else if cacheResult == nil {
		return cacheResult, nil
	}

	var result entity.AssessmentDetailView

	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get assessment detail: get from da failed",
			log.Err(err),
			log.String("id", "id"),
		)
		return nil, err
	}
	result = entity.AssessmentDetailView{
		ID:           assessment.ID,
		Title:        assessment.Title,
		ClassEndTime: assessment.ClassEndTime,
		ClassLength:  assessment.ClassLength,
		CompleteTime: assessment.CompleteTime,
		Status:       assessment.Status,
	}

	// fill attendances
	{
		attendanceIDs, err := da.GetAssessmentAttendanceDA().GetAttendanceIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get assessment attendances failed",
				log.Err(err),
				log.String("id", "id"),
			)
			return nil, err
		}
		studentService := external.GetStudentServiceProvider()
		students, err := studentService.BatchGet(ctx, attendanceIDs)
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get student failed",
				log.Err(err),
				log.String("id", id),
				log.Strings("attendance_ids", attendanceIDs),
			)
			return nil, err
		}
		for _, student := range students {
			result.Attendances = append(result.Attendances, &entity.AssessmentAttendanceView{
				ID:   student.ID,
				Name: student.Name,
			})
		}
	}

	// fill subject
	{
		nameMap, err := a.getSubjectNameMap(ctx, []string{assessment.SubjectID})
		if err != nil {
			log.Error(ctx, "get assessment detail: get subject name map failed",
				log.Err(err),
				log.String("id", id),
				log.String("subject_id", assessment.SubjectID),
			)
			return nil, err
		}
		result.Subject = entity.AssessmentSubject{
			ID:   assessment.SubjectID,
			Name: nameMap[assessment.SubjectID],
		}
	}

	// fill teacher
	{
		teacherIDs, err := assessment.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "get assessment detail: decode teacher ids failed",
				log.Err(err),
				log.String("id", id),
				log.Any("assessment", assessment),
			)
			return nil, err
		}
		teacherNameService := external.GetTeacherServiceProvider()
		items, err := teacherNameService.BatchGet(ctx, teacherIDs)
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get teacher failed",
				log.Err(err),
				log.Strings("teacher_ids", teacherIDs),
				log.Any("id", id),
			)
			return nil, err
		}
		for _, item := range items {
			result.Teachers = append(result.Teachers, &entity.AssessmentTeacher{
				ID:   item.ID,
				Name: item.Name,
			})
		}
	}

	// fill number of outcomes and activities
	{
		schedule, err := GetScheduleModel().GetPlainByID(ctx, assessment.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("id", id),
				log.Any("schedule_id", assessment.ScheduleID),
			)
			return nil, err
		}
		counts, err := GetContentModel().ContentDataCount(ctx, tx, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get number of activities and outcomes failed",
				log.Err(err),
				log.Any("id", id),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
		}
		result.NumberOfActivities = counts.SubContentCount
		result.NumberOfOutcomes = counts.OutcomesCount
	}

	// fill outcome attendance maps
	{
		outcomeMap := map[string]*entity.Outcome{}
		outcomeIDs, err := da.GetAssessmentOutcomeDA().GetOutcomeIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get outcome ids failed by assessment id",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		if len(outcomeIDs) > 0 {
			outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			for _, outcome := range outcomes {
				outcomeMap[outcome.ID] = outcome
			}
			sort.Sort(outcomeSliceSortByAssumedAndName(outcomes))
			outcomeAttendanceItems, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcome attendances failed by assessment id and outcome ids",
					log.Err(err),
					log.String("id", id),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			outcomeAttendanceMap := map[string][]string{}
			for _, item := range outcomeAttendanceItems {
				outcomeAttendanceMap[item.OutcomeID] = append(outcomeAttendanceMap[item.OutcomeID], item.AttendanceID)
			}
			assessmentOutcomeItems, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get assessment outcomes failed by assessment id and outcome ids",
					log.Err(err),
					log.String("id", id),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			assessmentOutcomeMap := map[string]entity.AssessmentOutcome{}
			for _, item := range assessmentOutcomeItems {
				assessmentOutcomeMap[item.OutcomeID] = *item
			}
			for _, outcome := range outcomes {
				newItem := entity.OutcomeAttendanceMapView{
					OutcomeID:     outcome.ID,
					OutcomeName:   outcome.Name,
					Assumed:       outcome.Assumed,
					AttendanceIDs: outcomeAttendanceMap[outcome.ID],
					Skip:          assessmentOutcomeMap[outcome.ID].Skip,
				}
				result.OutcomeAttendanceMaps = append(result.OutcomeAttendanceMaps, newItem)
			}
		}
	}

	if err := da.GetAssessmentRedisDA().CacheDetail(ctx, id, &result); err != nil {
		log.Info(ctx, "get assessment detail: cache failed",
			log.Err(err),
			log.String("id", "id"),
			log.Any("result", result),
		)
	}

	return &result, nil
}

func (a *assessmentModel) List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) (*entity.ListAssessmentsResult, error) {
	cacheResult, err := da.GetAssessmentRedisDA().List(ctx, cmd)
	if err != nil {
		log.Info(ctx, "list assessment: get cache list failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	} else if cacheResult != nil {
		return cacheResult, nil
	}

	cond := da.QueryAssessmentsCondition{
		Status:   cmd.Status,
		OrderBy:  cmd.OrderBy,
		Page:     cmd.Page,
		PageSize: cmd.PageSize,
	}
	{
		if cmd.TeacherName != nil {
			teacherService := external.GetTeacherServiceProvider()
			items, err := teacherService.Query(ctx, *cmd.TeacherName)
			if err != nil {
				log.Error(ctx, "list assessments: query teacher service failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return nil, err
			}
			if len(items) > 0 {
				if cond.TeacherIDs == nil {
					cond.TeacherIDs = &[]string{}
				}
				for _, item := range items {
					*cond.TeacherIDs = append(*cond.TeacherIDs, item.ID)
				}
			}
		}
	}

	var items []*entity.Assessment
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &items); err != nil {
		log.Error(ctx, "list assessments: query tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	total, err := da.GetAssessmentDA().CountTx(ctx, tx, &cond, &entity.Assessment{})
	if err != nil {
		log.Error(ctx, "list assessments: count tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	var subjectIDs, programIDs, teacherIDs []string

	for _, item := range items {
		subjectIDs = append(subjectIDs, item.SubjectID)
		programIDs = append(programIDs, item.ProgramID)
		teacherIDs, err := item.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "list assessment: decode teacher ids failed")
			return nil, err
		}
		teacherIDs = append(teacherIDs, teacherIDs...)
	}

	subjectNameMap, err := a.getSubjectNameMap(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "detail: get subject name map failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}

	programNameMap, err := a.getProgramNameMap(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "detail: get program name map failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}

	teacherNameMap, err := a.getTeacherNameMap(ctx, teacherIDs)
	if err != nil {
		log.Error(ctx, "detail: get teacher name map failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}

	result := entity.ListAssessmentsResult{Total: total}
	for _, item := range items {
		newItem := entity.AssessmentListView{
			ID:    item.ID,
			Title: item.Title,
			Subject: entity.AssessmentSubject{
				ID:   item.SubjectID,
				Name: subjectNameMap[item.SubjectID],
			},
			Program: entity.AssessmentProgram{
				ID:   item.ProgramID,
				Name: programNameMap[item.ProgramID],
			},
			ClassEndTime: item.ClassEndTime,
			CompleteTime: item.CompleteTime,
			Status:       item.Status,
		}
		teacherIDs, err := item.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "list assessment: decode teacher ids failed",
				log.Err(err),
				log.Any("item", item),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, teacherID := range teacherIDs {
			newItem.Teachers = append(newItem.Teachers, entity.AssessmentTeacher{
				ID:   teacherID,
				Name: teacherNameMap[teacherID],
			})
		}
		result.Items = append(result.Items, &newItem)
	}

	if err := da.GetAssessmentRedisDA().CacheList(ctx, cmd, &result); err != nil {
		log.Info(ctx, "list assessment: cache list failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("result", result),
		)
	}

	return &result, err
}

func (a *assessmentModel) getProgramNameMap(ctx context.Context, programIDs []string) (map[string]string, error) {
	programNameMap := map[string]string{}
	programService := external.GetProgramServiceProvider()
	items, err := programService.BatchGet(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get program failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}
	for _, item := range items {
		programNameMap[item.ID] = item.Name
	}
	return programNameMap, nil
}

func (a *assessmentModel) getSubjectNameMap(ctx context.Context, subjectIDs []string) (map[string]string, error) {
	subjectNameMap := map[string]string{}
	subjectService := external.GetSubjectServiceProvider()
	items, err := subjectService.BatchGet(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get subject failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}
	for _, item := range items {
		subjectNameMap[item.ID] = item.Name
	}
	return subjectNameMap, nil
}

func (a *assessmentModel) getTeacherNameMap(ctx context.Context, teacherIDs []string) (map[string]string, error) {
	teacherNameMap := map[string]string{}
	teacherNameService := external.GetTeacherServiceProvider()
	items, err := teacherNameService.BatchGet(ctx, teacherIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get teacher failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}
	for _, item := range items {
		teacherNameMap[item.ID] = item.Name
	}
	return teacherNameMap, nil
}

func (a *assessmentModel) getClassNameMap(ctx context.Context, classIDs []string) (map[string]string, error) {
	classNameMap := map[string]string{}
	classService := external.GetClassServiceProvider()
	items, err := classService.BatchGet(ctx, classIDs)
	if err != nil {
		log.Error(ctx, "get class name map: batch get class failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
		)
		return nil, err
	}
	for _, item := range items {
		classNameMap[item.ID] = item.Name
	}
	return classNameMap, nil
}

func (a *assessmentModel) Add(ctx context.Context, cmd entity.AddAssessmentCommand) (string, error) {
	var (
		outcomeIDs []string
		schedule   *entity.SchedulePlain
	)
	{
		var err error
		schedule, err = GetScheduleModel().GetPlainByID(ctx, cmd.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("schedule_id", cmd.ScheduleID),
			)
			switch err {
			case constant.ErrRecordNotFound, dbo.ErrRecordNotFound:
				return "", constant.ErrInvalidArgs
			default:
				return "", err
			}
		}
		outcomeIDs, err = GetContentModel().GetContentOutcomeByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get outcome failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
			return "", err
		}
	}

	var newID = utils.NewID()
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		{
			nowUnix := time.Now().Unix()
			newItem := entity.Assessment{
				ID:           newID,
				ScheduleID:   cmd.ScheduleID,
				ProgramID:    schedule.ProgramID,
				SubjectID:    schedule.SubjectID,
				ClassLength:  cmd.ClassLength,
				ClassEndTime: cmd.ClassEndTime,
				CreateAt:     nowUnix,
				UpdateAt:     nowUnix,
			}
			classNameMap, err := a.getClassNameMap(ctx, []string{schedule.ClassID})
			if err != nil {
				log.Error(ctx, "add assessment: get class name map failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
			newItem.Title = a.title(newItem.ClassEndTime, classNameMap[schedule.ClassID], schedule.Title)
			if err := newItem.EncodeAndSetTeacherIDs(schedule.TeacherIDs); err != nil {
				log.Error(ctx, "add assessment: encode and set teacher ids failed",
					log.Err(err),
					log.Strings("teacher_ids", schedule.TeacherIDs),
					log.Any("cmd", cmd),
				)
				return err
			}
			if len(outcomeIDs) == 0 {
				newItem.Status = entity.AssessmentStatusComplete
				newItem.CompleteTime = time.Now().Unix()
			} else {
				newItem.Status = entity.AssessmentStatusInProgress
			}
			if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newItem); err != nil {
				log.Error(ctx, "add assessment: add failed",
					log.Err(err),
					log.Any("cmd", cmd),
					log.Any("new_item", newItem),
				)
				return err
			}
		}
		{
			var items []*entity.AssessmentOutcome
			for _, outcomeID := range outcomeIDs {
				items = append(items, &entity.AssessmentOutcome{
					ID:           utils.NewID(),
					AssessmentID: newID,
					OutcomeID:    outcomeID,
					Skip:         false,
				})
			}
			if len(items) > 0 {
				if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, items); err != nil {
					log.Error(ctx, "add assessment: batch insert assessment outcome map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.Any("items", items),
					)
					return err
				}
			}
		}
		{
			outcomeMap := map[string]*entity.Outcome{}
			outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return err
			}
			for _, outcome := range outcomes {
				outcomeMap[outcome.ID] = outcome
			}

			var items []*entity.OutcomeAttendance
			for _, outcomeID := range outcomeIDs {
				if outcomeMap[outcomeID] == nil {
					continue
				}
				if !outcomeMap[outcomeID].Assumed {
					continue
				}
				for _, attendanceID := range cmd.AttendanceIDs {
					items = append(items, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: newID,
						OutcomeID:    outcomeID,
						AttendanceID: attendanceID,
					})
				}
			}
			if len(items) > 0 {
				if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
					log.Error(ctx, "add assessment: batch insert outcome attendance map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.Any("items", items),
					)
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	if err := da.GetAssessmentRedisDA().CleanList(ctx); err != nil {
		log.Info(ctx, "add assessment: clean list cache failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	}

	return newID, nil
}

func (a *assessmentModel) title(classEndTime int64, className string, lessonName string) string {
	return fmt.Sprintf("%s-%s-%s", time.Unix(classEndTime, 0).Format("20060102"), className, lessonName)
}

func (a *assessmentModel) Update(ctx context.Context, cmd entity.UpdateAssessmentCommand) error {
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, cmd.ID)
		if err != nil {
			log.Error(ctx, "update assessment: get assessment exclude soft deleted failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return err
		}
		if assessment.Status == entity.AssessmentStatusComplete {
			log.Info(ctx, "update assessment: assessment has completed, not allow update",
				log.Any("cmd", cmd),
			)
			return errors.New("update assessment: assessment has completed, not allow update")
		}
		if cmd.Action == entity.UpdateAssessmentActionComplete {
			if err := da.GetAssessmentDA().UpdateStatus(ctx, tx, cmd.ID, entity.AssessmentStatusComplete); err != nil {
				log.Error(ctx, "update assessment: update status to complete failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		if cmd.AttendanceIDs != nil {
			if err := da.GetAssessmentAttendanceDA().DeleteByAssessmentID(ctx, tx, cmd.ID); err != nil {
				log.Error(ctx, "update assessment: delete assessment attendance failed by assessment id",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
			var items []*entity.AssessmentAttendance
			for _, attendanceID := range *cmd.AttendanceIDs {
				items = append(items, &entity.AssessmentAttendance{
					ID:           utils.NewID(),
					AssessmentID: cmd.ID,
					AttendanceID: attendanceID,
				})
			}
			if err := da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "update assessment: batch insert assessment attendance map failed",
					log.Err(err),
					log.Any("items", items),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		if cmd.OutcomeAttendanceMaps != nil {
			var outcomeIDs []string
			for _, item := range *cmd.OutcomeAttendanceMaps {
				outcomeIDs = append(outcomeIDs, item.OutcomeID)
			}
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, cmd.ID, outcomeIDs); err != nil {
				log.Error(ctx, "update assessment: batch delete outcome attendance map failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
					log.Any("cmd", cmd),
				)
				return err
			}
			var items []*entity.OutcomeAttendance
			for _, item := range *cmd.OutcomeAttendanceMaps {
				for _, attendanceID := range item.AttendanceIDs {
					items = append(items, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: cmd.ID,
						OutcomeID:    item.OutcomeID,
						AttendanceID: attendanceID,
					})
				}
				if err := da.GetAssessmentOutcomeDA().UpdateSkipField(ctx, tx, cmd.ID, item.OutcomeID, item.Skip); err != nil {
					log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.String("id", cmd.ID),
						log.String("outcome_id", item.OutcomeID),
						log.Bool("skip", item.Skip),
					)
				}
			}
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("items", items),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update assessment: tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return err
	}

	if err := da.GetAssessmentRedisDA().CleanDetail(ctx, cmd.ID); err != nil {
		log.Info(ctx, "update assessment: clean detail cache failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	}
	if err := da.GetAssessmentRedisDA().CleanList(ctx); err != nil {
		log.Info(ctx, "update assessment: clean list cache failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
	}

	return nil
}

func (a *assessmentModel) existsTeachersByIDs(ctx context.Context, ids []string) (bool, error) {
	teacherService := external.GetTeacherServiceProvider()
	if _, err := teacherService.BatchGet(ctx, ids); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check teacher exists: not found teachers",
				log.Err(err),
				log.Strings("ids", ids),
			)
			return false, nil
		default:
			log.Error(ctx, "check teacher exists: batch get teachers failed",
				log.Err(err),
				log.Strings("ids", ids),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsSubjectByID(ctx context.Context, id string) (bool, error) {
	subjectService := external.GetSubjectServiceProvider()
	if _, err := subjectService.BatchGet(ctx, []string{id}); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check subject exists: not found subjects",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check subject exists: batch get subjects failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsProgramByID(ctx context.Context, id string) (bool, error) {
	programService := external.GetProgramServiceProvider()
	if _, err := programService.BatchGet(ctx, []string{id}); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check program exists: not found programs",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check program exists: batch get programs failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}
